package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"pi-gpio-dashboard/automation"
	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/db"
	"pi-gpio-dashboard/master"
	"pi-gpio-dashboard/node"
	"pi-gpio-dashboard/sensors"
)

// Handler holds all API dependencies
type Handler struct {
	cfg           *config.Config
	manager       *node.Manager
	db            *db.DB
	hub           *WebSocketHub
	agentPool     *master.AgentPool
	engine        *automation.Engine
	sensorManager *sensors.SensorManager
}

// NewHandler creates a new API handler
func NewHandler(cfg *config.Config, manager *node.Manager, database *db.DB, hub *WebSocketHub, agentPool *master.AgentPool, engine *automation.Engine, sensorManager *sensors.SensorManager) *Handler {
	return &Handler{
		cfg:           cfg,
		manager:       manager,
		db:            database,
		hub:           hub,
		agentPool:     agentPool,
		engine:        engine,
		sensorManager: sensorManager,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Node management
		api.GET("/nodes", h.GetNodes)
		api.GET("/nodes/:id/pins", h.GetNodePins)
		api.GET("/nodes/:id/pins/:pin/status", h.GetPinStatus)
		api.POST("/nodes/:id/pins/:pin/action", h.ExecuteAction)

		// Automations
		api.GET("/automations", h.GetAutomations)
		api.POST("/automations", h.CreateAutomation)
		api.GET("/automations/:id", h.GetAutomation)
		api.PUT("/automations/:id", h.UpdateAutomation)
		api.DELETE("/automations/:id", h.DeleteAutomation)
		api.PUT("/automations/:id/enable", h.EnableAutomation)

		// Logs
		api.GET("/logs", h.GetLogs)

		// Sensors
		api.GET("/sensors", h.GetSensors)
		api.GET("/sensors/:id/read", h.ReadSensor)
		api.POST("/sensors/:id/enable", h.EnableSensor)
		api.POST("/sensors/:id/disable", h.DisableSensor)

		// Agent registration
		api.POST("/agents/register", h.RegisterAgent)
		api.DELETE("/agents/:id", h.UnregisterAgent)
		api.POST("/agents/:id/heartbeat", h.AgentHeartbeat)
		api.GET("/agents", h.GetAgents)
	}

	// WebSocket
	r.GET("/ws", h.hub.HandleWebSocket)
}

// GetNodes returns all nodes including local and agents
func (h *Handler) GetNodes(c *gin.Context) {
	nodes := make([]interface{}, 0)

	// Add local node
	localState := h.manager.GetNodeState()
	nodes = append(nodes, localState)

	// Add agent nodes if pool exists
	if h.agentPool != nil {
		for nodeID, agent := range h.agentPool.GetAllAgents() {
			info := agent.GetNodeInfo()
			state := agent.GetState()
			
			// Extract pins from state (state_full contains full NodeState, we need just pins)
			pins := state
			if pinsData, ok := state["pins"]; ok {
				if pinsMap, ok := pinsData.(map[string]interface{}); ok {
					pins = pinsMap
				}
			}
			
			// Create node state from agent data
			agentNode := map[string]interface{}{
				"id":     nodeID,
				"name":   info["name"],
				"role":   "agent",
				"status": func() string {
					if agent.IsConnected() {
						return "online"
					}
					return "offline"
				}(),
				"ip":           info["ip"],
				"pins":         pins,
			}

			// Pass through mock_gpio flag from agent state if available
			if mockGPIO, ok := state["mock_gpio"]; ok {
				agentNode["mock_gpio"] = mockGPIO
			}
			nodes = append(nodes, agentNode)
		}
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// GetNodePins returns all pins for a node (local or agent)
func (h *Handler) GetNodePins(c *gin.Context) {
	nodeID := c.Param("id")

	if nodeID == h.cfg.Node.ID {
		// Local node
		pins := h.manager.GetAllStates()
		c.JSON(http.StatusOK, gin.H{"pins": pins})
		return
	}

	// Agent node
	if h.agentPool != nil {
		if agent, ok := h.agentPool.GetAgent(nodeID); ok && agent.IsConnected() {
			pins, err := agent.GetPins()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, pins)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
}

// GetPinStatus returns the status of a specific pin (local or agent)
func (h *Handler) GetPinStatus(c *gin.Context) {
	nodeID := c.Param("id")
	pinID := c.Param("pin")

	if nodeID == h.cfg.Node.ID {
		// Local node
		state, err := h.manager.GetState(pinID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, state)
		return
	}

	// Agent node
	if h.agentPool != nil {
		if agent, ok := h.agentPool.GetAgent(nodeID); ok && agent.IsConnected() {
			// Forward request to agent
			// For simplicity, return the cached state
			state := agent.GetState()
			if pinState, ok := state[pinID]; ok {
				c.JSON(http.StatusOK, pinState)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "pin not found"})
}

// ExecuteActionRequest represents an action request
type ExecuteActionRequest struct {
	Action string                 `json:"action" binding:"required"`
	Params map[string]interface{} `json:"params"`
}

// ExecuteAction performs an action on a pin (local or agent)
func (h *Handler) ExecuteAction(c *gin.Context) {
	nodeID := c.Param("id")
	pinID := c.Param("pin")

	var req ExecuteActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var result interface{}
	var newState interface{}
	var err error

	if nodeID == h.cfg.Node.ID {
		// Local execution
		result, err = h.manager.ExecuteAction(pinID, req.Action, req.Params)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		state, _ := h.manager.GetState(pinID)
		newState = state.State
	} else if h.agentPool != nil {
		// Forward to agent
		err = h.agentPool.ForwardAction(nodeID, pinID, req.Action, req.Params)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result = true
		// Get cached state from agent
		if agent, ok := h.agentPool.GetAgent(nodeID); ok {
			state := agent.GetState()
			if pinState, ok := state[pinID]; ok {
				newState = pinState
			}
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	// Log the action
	paramsJSON, _ := json.Marshal(req.Params)
	resultStr, _ := json.Marshal(result)
	h.db.LogAction(nodeID, pinID, req.Action, string(paramsJSON), string(resultStr), "user")
	h.db.TrimActionLogs(1000)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"result":     result,
		"new_state":  newState,
	})
}

// GetAutomations returns all automation rules
func (h *Handler) GetAutomations(c *gin.Context) {
	automations, err := h.db.GetAutomations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse configs
	result := make([]map[string]interface{}, 0, len(automations))
	for _, auto := range automations {
		var config struct {
			Trigger config.TriggerConfig     `json:"trigger"`
			Actions []config.ActionStepConfig `json:"actions"`
		}
		json.Unmarshal([]byte(auto.Config), &config)

		result = append(result, map[string]interface{}{
			"id":        auto.ID,
			"name":      auto.Name,
			"enabled":   auto.Enabled,
			"trigger":   config.Trigger,
			"actions":   config.Actions,
			"created_at": auto.CreatedAt,
			"updated_at": auto.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"automations": result})
}

// CreateAutomationRequest represents the request body for creating automation
type CreateAutomationRequest struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name" binding:"required"`
	Enabled   bool                     `json:"enabled"`
	Trigger   config.TriggerConfig     `json:"trigger" binding:"required"`
	Actions   []config.ActionStepConfig `json:"actions" binding:"required"`
}

// CreateAutomation creates a new automation rule
func (h *Handler) CreateAutomation(c *gin.Context) {
	var req CreateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Serialize config
	config := struct {
		Trigger config.TriggerConfig     `json:"trigger"`
		Actions []config.ActionStepConfig `json:"actions"`
	}{
		Trigger: req.Trigger,
		Actions: req.Actions,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to serialize config"})
		return
	}

	id := req.ID
	if id == "" {
		id = fmt.Sprintf("auto-%d", time.Now().UnixNano())
	}

	automation := &db.Automation{
		ID:      id,
		Name:    req.Name,
		Enabled: req.Enabled,
		Config:  string(configJSON),
	}

	if err := h.db.SaveAutomation(automation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"name":    req.Name,
		"enabled": req.Enabled,
		"trigger": req.Trigger,
		"actions": req.Actions,
	})
}

// GetAutomation returns a single automation rule
func (h *Handler) GetAutomation(c *gin.Context) {
	id := c.Param("id")
	automation, err := h.db.GetAutomation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "automation not found"})
		return
	}

	var config struct {
		Trigger config.TriggerConfig     `json:"trigger"`
		Actions []config.ActionStepConfig `json:"actions"`
	}
	json.Unmarshal([]byte(automation.Config), &config)

	c.JSON(http.StatusOK, gin.H{
		"id":        automation.ID,
		"name":      automation.Name,
		"enabled":   automation.Enabled,
		"trigger":   config.Trigger,
		"actions":   config.Actions,
		"created_at": automation.CreatedAt,
		"updated_at": automation.UpdatedAt,
	})
}

// UpdateAutomationRequest represents the request body for updating automation
type UpdateAutomationRequest struct {
	Name    string                   `json:"name"`
	Enabled *bool                    `json:"enabled,omitempty"`
	Trigger *config.TriggerConfig     `json:"trigger,omitempty"`
	Actions []config.ActionStepConfig `json:"actions,omitempty"`
}

// UpdateAutomation updates an automation rule
func (h *Handler) UpdateAutomation(c *gin.Context) {
	id := c.Param("id")

	// Get existing automation
	existing, err := h.db.GetAutomation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "automation not found"})
		return
	}

	var req UpdateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse existing config
	var existingConfig struct {
		Trigger config.TriggerConfig     `json:"trigger"`
		Actions []config.ActionStepConfig `json:"actions"`
	}
	if err := json.Unmarshal([]byte(existing.Config), &existingConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse existing config"})
		return
	}

	// Update fields
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.Trigger != nil {
		existingConfig.Trigger = *req.Trigger
	}
	if req.Actions != nil {
		existingConfig.Actions = req.Actions
	}

	// Serialize updated config
	configJSON, err := json.Marshal(existingConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to serialize config"})
		return
	}
	existing.Config = string(configJSON)

	if err := h.db.SaveAutomation(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"name":    existing.Name,
		"enabled": existing.Enabled,
		"trigger": existingConfig.Trigger,
		"actions": existingConfig.Actions,
	})
}

// DeleteAutomation removes an automation rule
func (h *Handler) DeleteAutomation(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.DeleteAutomation(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// EnableAutomation enables or disables an automation rule
func (h *Handler) EnableAutomation(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	automation, err := h.db.GetAutomation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "automation not found"})
		return
	}

	automation.Enabled = req.Enabled
	if err := h.db.SaveAutomation(automation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Also update engine
	if h.engine != nil {
		h.engine.EnableRule(id, req.Enabled)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"enabled": req.Enabled,
	})
}

// GetLogs returns action logs
func (h *Handler) GetLogs(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		// Parse limit, default to 50
		// (simplified, in production use strconv.Atoi)
	}
	nodeID := c.Query("node")

	logs, err := h.db.GetActionLogs(limit, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// AgentRegistration represents an agent registration request
type AgentRegistration struct {
	NodeID   string `json:"node_id" binding:"required"`
	NodeName string `json:"node_name" binding:"required"`
	IP       string `json:"ip" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
}

// RegisterAgent handles agent registration
func (h *Handler) RegisterAgent(c *gin.Context) {
	var req AgentRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify API key
	if req.APIKey != h.cfg.Security.APIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	// Add agent to pool
	agent := h.agentPool.AddAgent(req.NodeID, req.NodeName, req.IP, req.APIKey, req.Port)
	
	// Set up state change forwarding
	h.agentPool.OnAgentStateChange(func(nodeID string, data map[string]interface{}) {
		h.hub.Broadcast(data)
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent registered",
		"agent": agent.GetNodeInfo(),
	})
}

// UnregisterAgent removes an agent from the pool
func (h *Handler) UnregisterAgent(c *gin.Context) {
	nodeID := c.Param("id")
	
	if agent, ok := h.agentPool.GetAgent(nodeID); ok {
		agent.Disconnect()
		h.agentPool.RemoveAgent(nodeID)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Agent unregistered",
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
	}
}

// AgentHeartbeat handles agent heartbeat
func (h *Handler) AgentHeartbeat(c *gin.Context) {
	nodeID := c.Param("id")
	
	var req AgentRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify API key
	if req.APIKey != h.cfg.Security.APIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	agent, ok := h.agentPool.GetAgent(nodeID)
	if !ok {
		// Agent not found, auto-register it
		agent = h.agentPool.AddAgent(req.NodeID, req.NodeName, req.IP, req.APIKey, req.Port)
	}

	// Update last seen
	agent.UpdateLastSeen()

	// Try to connect if not connected
	if !agent.IsConnected() {
		go agent.Connect()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Heartbeat received",
	})
}

// GetAgents returns all registered agents
func (h *Handler) GetAgents(c *gin.Context) {
	agents := h.agentPool.GetAllAgents()
	
	result := make([]map[string]interface{}, 0, len(agents))
	for _, agent := range agents {
		info := agent.GetNodeInfo()
		info["health"] = agent.GetHealth()
		result = append(result, info)
	}

	c.JSON(http.StatusOK, gin.H{"agents": result})
}

// GetSensors returns all registered sensors
func (h *Handler) GetSensors(c *gin.Context) {
	if h.sensorManager == nil {
		c.JSON(http.StatusOK, gin.H{"sensors": []interface{}{}})
		return
	}

	sensors := h.sensorManager.GetAllSensors()
	result := make([]map[string]interface{}, 0, len(sensors))
	for id, sensor := range sensors {
		reading, err := sensor.Read()
		
		sensorData := map[string]interface{}{
			"id":   id,
			"type": sensor.Type(),
			"reading": map[string]interface{}{
				"value":     reading.Value,
				"unit":      reading.Unit,
				"timestamp": reading.Timestamp,
				"metadata":  reading.Metadata,
			},
		}
		
		if err != nil {
			sensorData["error"] = err.Error()
		}
		
		result = append(result, sensorData)
	}

	c.JSON(http.StatusOK, gin.H{"sensors": result})
}

// ReadSensor reads a specific sensor
func (h *Handler) ReadSensor(c *gin.Context) {
	id := c.Param("id")

	if h.sensorManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sensor manager not initialized"})
		return
	}

	sensor, exists := h.sensorManager.GetSensor(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	reading, err := sensor.Read()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   id,
		"type": sensor.Type(),
		"reading": map[string]interface{}{
			"value":     reading.Value,
			"unit":      reading.Unit,
			"timestamp": reading.Timestamp,
			"metadata":  reading.Metadata,
		},
	})
}

// EnableSensor enables a sensor
func (h *Handler) EnableSensor(c *gin.Context) {
	id := c.Param("id")

	if h.sensorManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sensor manager not initialized"})
		return
	}

	sensor, exists := h.sensorManager.GetSensor(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	if err := sensor.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Sensor enabled"})
}

// DisableSensor disables a sensor
func (h *Handler) DisableSensor(c *gin.Context) {
	id := c.Param("id")

	if h.sensorManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sensor manager not initialized"})
		return
	}

	sensor, exists := h.sensorManager.GetSensor(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	if err := sensor.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Sensor disabled"})
}