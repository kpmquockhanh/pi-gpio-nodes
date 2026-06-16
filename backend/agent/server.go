package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/node"
)

// AgentServer runs on each Pi and handles local GPIO
// It is a thin relay: no local DB, no frontend, no automation engine.
type AgentServer struct {
	cfg       *config.Config
	manager   *node.Manager
	masterWS  *websocket.Conn
	router    *gin.Engine
	port      int
	apiKey    string
}

// NewAgentServer creates a new agent server
func NewAgentServer(cfg *config.Config, manager *node.Manager) *AgentServer {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	return &AgentServer{
		cfg:     cfg,
		manager: manager,
		router:  router,
		port:    cfg.Network.ListenPort,
		apiKey:  cfg.Security.APIKey,
	}
}

// SetupRoutes configures all agent routes
func (a *AgentServer) SetupRoutes() {
	// Health check
	a.router.GET("/api/health", a.handleHealth)

	// Pin routes
	a.router.GET("/api/pins", a.handleGetPins)
	a.router.GET("/api/pins/:pin/status", a.handleGetPinStatus)
	a.router.POST("/api/pins/:pin/action", a.handleExecuteAction)

	// WebSocket for master connection
	a.router.GET("/ws", a.handleWebSocket)
}

// Run starts the agent server
func (a *AgentServer) Run() error {
	addr := fmt.Sprintf(":%d", a.port)
	log.Printf("Agent server listening on %s", addr)
	return a.router.Run(addr)
}

// ConnectToMaster establishes WebSocket connection to master
func (a *AgentServer) ConnectToMaster() {
	masterURL := fmt.Sprintf("ws://%s:%d/ws?api_key=%s", 
		a.cfg.Network.MasterNode, 
		8080, // default master port
		a.apiKey)

	for {
		conn, _, err := websocket.DefaultDialer.Dial(masterURL, nil)
		if err != nil {
			log.Printf("Failed to connect to master: %v, retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Connected to master")
		a.masterWS = conn

		// Send initial state
		a.broadcastState()

		// Handle incoming messages from master
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Master connection lost: %v", err)
				a.masterWS = nil
				conn.Close()
				break
			}

			// Process commands from master
			var cmd MasterCommand
			if err := json.Unmarshal(msg, &cmd); err == nil {
				a.handleMasterCommand(cmd)
			}
		}

		// Reconnect
		time.Sleep(5 * time.Second)
	}
}

// MasterCommand represents a command from master
type MasterCommand struct {
	Type   string                 `json:"type"`
	PinID  string                 `json:"pin_id"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

func (a *AgentServer) handleMasterCommand(cmd MasterCommand) interface{} {
	switch cmd.Type {
	case "action":
		result, err := a.manager.ExecuteAction(cmd.PinID, cmd.Action, cmd.Params)
		if err != nil {
			log.Printf("Failed to execute command: %v", err)
			return nil
		}
		return result
	}
	return nil
}

// Broadcast state change to master
func (a *AgentServer) broadcastState() {
	if a.masterWS == nil {
		return
	}

	state := a.manager.GetNodeState()
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "state_full",
		"node":  a.cfg.Node.ID,
		"state": state,
	})

	a.masterWS.WriteMessage(websocket.TextMessage, data)
}

// BroadcastPinUpdate sends a single pin update to master
func (a *AgentServer) BroadcastPinUpdate(pinID string, state interface{}) {
	if a.masterWS == nil {
		return
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type":  "state_update",
		"node":  a.cfg.Node.ID,
		"pin":   pinID,
		"state": state,
	})

	a.masterWS.WriteMessage(websocket.TextMessage, data)
}

// HTTP Handlers

func (a *AgentServer) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"node_id":   a.cfg.Node.ID,
		"node_name": a.cfg.Node.Name,
		"uptime":    time.Now().Unix(),
	})
}

func (a *AgentServer) handleGetPins(c *gin.Context) {
	pins := a.manager.GetAllStates()
	c.JSON(http.StatusOK, gin.H{"pins": pins})
}

func (a *AgentServer) handleGetPinStatus(c *gin.Context) {
	pinID := c.Param("pin")
	state, err := a.manager.GetState(pinID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

func (a *AgentServer) handleExecuteAction(c *gin.Context) {
	pinID := c.Param("pin")

	var req struct {
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.manager.ExecuteAction(pinID, req.Action, req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast state change to master (master is the single source of truth for logs)
	state, _ := a.manager.GetState(pinID)
	a.BroadcastPinUpdate(pinID, state.State)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"result":    result,
		"new_state": state.State,
	})
}

func (a *AgentServer) handleWebSocket(c *gin.Context) {
	// This WebSocket is for master to connect to agent
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Handle master WebSocket connection
	log.Println("Master connected to agent WebSocket")

	// Send initial state
	a.sendStateToMaster(conn)

	// Setup state change broadcasting via this connection
	for _, pin := range a.manager.GetAllStates() {
		pinID := pin.ID
		a.manager.OnStateChange(pinID, func(id string, state interface{}) {
			a.sendPinUpdateToMaster(conn, id, state)
		})
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Master WebSocket disconnected")
			return
		}

		var cmd MasterCommand
		if err := json.Unmarshal(msg, &cmd); err == nil {
			result := a.handleMasterCommand(cmd)
			if result != nil {
				// Send state update after command
				state, _ := a.manager.GetState(cmd.PinID)
				a.sendPinUpdateToMaster(conn, cmd.PinID, state.State)
			}
		}
	}
}

func (a *AgentServer) sendStateToMaster(conn *websocket.Conn) {
	state := a.manager.GetNodeState()
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "state_full",
		"node":  a.cfg.Node.ID,
		"state": state,
	})
	conn.WriteMessage(websocket.TextMessage, data)
}

func (a *AgentServer) sendPinUpdateToMaster(conn *websocket.Conn, pinID string, state interface{}) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "state_update",
		"node":  a.cfg.Node.ID,
		"pin":   pinID,
		"state": state,
	})
	conn.WriteMessage(websocket.TextMessage, data)
}