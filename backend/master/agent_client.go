package master

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// HealthStatus represents the health status of an agent
type HealthStatus string

const (
	HealthHealthy    HealthStatus = "healthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthOffline   HealthStatus = "offline"
)

// AgentHealth represents agent health information
type AgentHealth struct {
	Status      HealthStatus `json:"status"`
	Latency     int64        `json:"latency_ms"`
	LastSeen    time.Time    `json:"last_seen"`
	LastError   string       `json:"last_error,omitempty"`
	Connected   bool         `json:"connected"`
	Uptime      int64        `json:"uptime_seconds"`
}

// AgentClient represents a connection to an agent Pi
type AgentClient struct {
	nodeID      string
	nodeName    string
	ip string
	port        int
	apiKey      string
	conn        *websocket.Conn
	mu          sync.RWMutex
	connected   bool
	lastSeen    time.Time
	state       map[string]interface{}
	handlers    []func(string, map[string]interface{})
	health      AgentHealth
	lastError   string
	connectedAt time.Time
}

// NewAgentClient creates a new agent client
func NewAgentClient(nodeID, nodeName, ip, apiKey string, port int) *AgentClient {
	return &AgentClient{
		nodeID:   nodeID,
		nodeName: nodeName,
		ip:       ip,
		port:     port,
		apiKey:   apiKey,
		state:    make(map[string]interface{}),
		handlers: make([]func(string, map[string]interface{}), 0),
	}
}

// Connect establishes WebSocket connection to agent
func (c *AgentClient) Connect() error {
	wsURL := fmt.Sprintf("ws://%s:%d/ws?api_key=%s", c.ip, c.port, c.apiKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to agent %s: %w", c.nodeID, err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.lastSeen = time.Now()
	c.connectedAt = time.Now()
	c.health = AgentHealth{Status: HealthHealthy, Connected: true}
	c.mu.Unlock()

	log.Printf("Connected to agent: %s (%s)", c.nodeName, c.nodeID)

	// Proactively fetch pins via HTTP so we have state even before WS messages arrive
	if pinsResult, err := c.GetPins(); err == nil {
		if pinsData, ok := pinsResult["pins"]; ok {
			if pinsMap, ok := pinsData.(map[string]interface{}); ok {
				c.mu.Lock()
				c.state["pins"] = pinsMap
				c.mu.Unlock()
				log.Printf("Fetched %d pins from agent %s via HTTP", len(pinsMap), c.nodeID)
			}
		}
	} else {
		log.Printf("Failed to fetch pins from agent %s via HTTP: %v", c.nodeID, err)
	}

	// Start message handler
	go c.readLoop()

	return nil
}

// Disconnect closes the connection
func (c *AgentClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
}

// IsConnected returns connection status
func (c *AgentClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetNodeInfo returns node information
func (c *AgentClient) GetNodeInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":            c.nodeID,
		"name":          c.nodeName,
		"ip":            c.ip,
		"port":          c.port,
		"connected":     c.IsConnected(),
		"last_seen":     c.lastSeen,
	}
}

// SendCommand sends an action command to the agent
func (c *AgentClient) SendCommand(pinID, action string, params map[string]interface{}) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected to agent %s", c.nodeID)
	}

	cmd := map[string]interface{}{
		"type":   "action",
		"pin_id": pinID,
		"action": action,
		"params": params,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// GetState returns the cached agent state
func (c *AgentClient) GetState() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state := make(map[string]interface{})
	for k, v := range c.state {
		state[k] = v
	}
	return state
}

// OnStateChange registers a handler for state changes
func (c *AgentClient) OnStateChange(handler func(string, map[string]interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *AgentClient) readLoop() {
	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Agent %s connection lost: %v", c.nodeID, err)
			c.mu.Lock()
			c.connected = false
			c.conn = nil
			c.mu.Unlock()
			return
		}

		c.mu.Lock()
		c.lastSeen = time.Now()
		c.mu.Unlock()

		// Parse message
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err != nil {
			continue
		}

		msgType, _ := data["type"].(string)

		c.mu.Lock()
		switch msgType {
		case "state_full":
			if state, ok := data["state"].(map[string]interface{}); ok {
				c.state = state
				log.Printf("Received state_full from agent %s with %d top-level keys", c.nodeID, len(state))
				if pins, ok := state["pins"]; ok {
					if pinsMap, ok := pins.(map[string]interface{}); ok {
						log.Printf("Agent %s has %d pins in state_full", c.nodeID, len(pinsMap))
					}
				}
			} else {
				log.Printf("Received state_full from agent %s but state was not a map", c.nodeID)
			}
		case "state_update":
			pinID, _ := data["pin"].(string)
			state, _ := data["state"]
			c.state[pinID] = state
			log.Printf("Received state_update from agent %s for pin %s", c.nodeID, pinID)
		}

		// Notify handlers
		for _, handler := range c.handlers {
			go handler(c.nodeID, data)
		}
		c.mu.Unlock()
	}
}

// UpdateLastSeen updates the last seen timestamp
func (c *AgentClient) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastSeen = time.Now()
}

// GetHealth returns the agent's current health status
func (c *AgentClient) GetHealth() AgentHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Calculate health status
	status := HealthHealthy
	if !c.connected {
		status = HealthOffline
	} else if time.Since(c.lastSeen) > 30*time.Second {
		status = HealthDegraded
	}

	uptime := int64(0)
	if !c.connectedAt.IsZero() {
		uptime = int64(time.Since(c.connectedAt).Seconds())
	}

	return AgentHealth{
		Status:    status,
		LastSeen:  c.lastSeen,
		LastError: c.lastError,
		Connected: c.connected,
		Uptime:    uptime,
	}
}

// CheckHealth performs an active health check
func (c *AgentClient) CheckHealth() error {
	start := time.Now()
	url := fmt.Sprintf("http://%s:%d/api/health", c.ip, c.port)
	resp, err := http.Get(url)
	if err != nil {
		c.mu.Lock()
		c.lastError = err.Error()
		c.mu.Unlock()
		return err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		c.mu.Lock()
		c.lastError = err.Error()
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.lastSeen = time.Now()
	c.health = AgentHealth{
		Status:   HealthHealthy,
		Latency:  latency,
		LastSeen: c.lastSeen,
		Connected: c.connected,
	}
	c.mu.Unlock()

	return nil
}

// GetPins fetches pin states from agent via HTTP
func (c *AgentClient) GetPins() (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d/api/pins", c.ip, c.port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}