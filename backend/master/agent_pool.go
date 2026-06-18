package master

import (
	"fmt"
	"log"
	"sync"
	"time"

	"pi-gpio-dashboard/config"
)

// AgentPool manages connections to all agent nodes
type AgentPool struct {
	mu      sync.RWMutex
	agents  map[string]*AgentClient
	config  *config.Config
}

// NewAgentPool creates a new agent pool
func NewAgentPool(cfg *config.Config) *AgentPool {
	return &AgentPool{
		agents: make(map[string]*AgentClient),
		config: cfg,
	}
}

// LoadAgentsFromConfig creates agent clients from config automations
func (p *AgentPool) LoadAgentsFromConfig() {
	// In a real implementation, you'd load from a separate agents config
	// For now, we'll just initialize the pool
	log.Println("Agent pool initialized")
}

// AddAgent adds an agent to the pool
func (p *AgentPool) AddAgent(nodeID, nodeName, ip, apiKey string, port int) *AgentClient {
	p.mu.Lock()
	defer p.mu.Unlock()

	client := NewAgentClient(nodeID, nodeName, ip, apiKey, port)
	p.agents[nodeID] = client
	
	// Try to connect
	go func() {
		if err := client.Connect(); err != nil {
			log.Printf("Failed to connect to agent %s: %v", nodeID, err)
		}
	}()

	return client
}

// RemoveAgent removes an agent from the pool
func (p *AgentPool) RemoveAgent(nodeID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if agent, ok := p.agents[nodeID]; ok {
		agent.Disconnect()
		delete(p.agents, nodeID)
	}
}

// GetAgent returns an agent client by ID
func (p *AgentPool) GetAgent(nodeID string) (*AgentClient, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, ok := p.agents[nodeID]
	return agent, ok
}

// GetAllAgents returns all agent clients
func (p *AgentPool) GetAllAgents() map[string]*AgentClient {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := make(map[string]*AgentClient)
	for k, v := range p.agents {
		agents[k] = v
	}
	return agents
}

// GetAllNodeStates returns states from all connected agents
func (p *AgentPool) GetAllNodeStates() []map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	states := make([]map[string]interface{}, 0)
	for _, agent := range p.agents {
		if agent.IsConnected() {
			state := agent.GetState()
			if len(state) > 0 {
				states = append(states, state)
			}
		}
	}
	return states
}

// GetNodeState returns a specific node's state
func (p *AgentPool) GetNodeState(nodeID string) (map[string]interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, ok := p.agents[nodeID]
	if !ok || !agent.IsConnected() {
		return nil, false
	}

	return agent.GetState(), true
}

// ForwardAction sends an action to a specific agent
func (p *AgentPool) ForwardAction(nodeID, pinID, action string, params map[string]interface{}) error {
	agent, ok := p.GetAgent(nodeID)
	if !ok {
		return fmt.Errorf("agent %s not found", nodeID)
	}

	if !agent.IsConnected() {
		return fmt.Errorf("agent %s not connected", nodeID)
	}

	return agent.SendCommand(pinID, action, params)
}

// StartHealthChecks begins periodic health checks for all agents
func (p *AgentPool) StartHealthChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			p.checkAllAgents()
		}
	}()
}

func (p *AgentPool) checkAllAgents() {
	p.mu.RLock()
	agents := make([]*AgentClient, 0, len(p.agents))
	for _, agent := range p.agents {
		agents = append(agents, agent)
	}
	p.mu.RUnlock()

	for _, agent := range agents {
		// Check health via HTTP
		if err := agent.CheckHealth(); err != nil {
			log.Printf("Health check failed for agent %s: %v", agent.nodeID, err)
			
			// If health check fails, try to reconnect WebSocket
			if !agent.IsConnected() {
				log.Printf("Attempting to reconnect to agent: %s", agent.nodeID)
				if err := agent.Connect(); err != nil {
					log.Printf("Reconnection failed for agent %s: %v", agent.nodeID, err)
				}
			}
		} else {
			// If health check succeeds but WebSocket is not connected, reconnect
			if !agent.IsConnected() {
				log.Printf("Agent %s healthy but WebSocket disconnected, reconnecting...", agent.nodeID)
				go agent.Connect()
			}
		}
	}
}

// GetNodeStatus returns connection status for all agents
func (p *AgentPool) GetNodeStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := make(map[string]interface{})
	for nodeID, agent := range p.agents {
		status[nodeID] = map[string]interface{}{
			"connected":  agent.IsConnected(),
			"last_seen":  agent.lastSeen,
			"node_name":  agent.nodeName,
		}
	}
	return status
}

// GetConnectedCount returns number of connected agents
func (p *AgentPool) GetConnectedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, agent := range p.agents {
		if agent.IsConnected() {
			count++
		}
	}
	return count
}

// OnAgentStateChange registers a callback for all agents
func (p *AgentPool) OnAgentStateChange(handler func(nodeID string, data map[string]interface{})) {
	p.mu.RLock()
	agents := make([]*AgentClient, 0, len(p.agents))
	for _, agent := range p.agents {
		agents = append(agents, agent)
	}
	p.mu.RUnlock()

	for _, agent := range agents {
		agent.OnStateChange(handler)
	}
}