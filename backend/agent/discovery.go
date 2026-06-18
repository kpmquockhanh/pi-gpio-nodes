package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"pi-gpio-dashboard/config"
)

// Discovery handles agent registration with master
// Agent auto-discovery allows agents to register themselves when they come online
// without manual configuration on the master
//
// Usage:
//   d := agent.NewDiscovery(cfg)
//   d.RegisterWithMaster() // Attempts to register with master
//   d.StartHeartbeat()   // Starts sending periodic heartbeats
//
// The agent will periodically retry registration if it fails
// and will send heartbeats to maintain connection status

type Discovery struct {
	cfg           *config.Config
	masterURL     string
	registered    bool
	stopHeartbeat chan bool
}

// NewDiscovery creates a new discovery service
func NewDiscovery(cfg *config.Config) *Discovery {
	masterURL := fmt.Sprintf("http://%s:%d", cfg.Network.MasterNode, 8080)
	return &Discovery{
		cfg:           cfg,
		masterURL:     masterURL,
		stopHeartbeat: make(chan bool),
	}
}

// AgentInfo represents agent registration information
type AgentInfo struct {
	NodeID      string `json:"node_id"`
	NodeName    string `json:"node_name"`
	IP string `json:"ip"`
	Port        int    `json:"port"`
	APIKey      string `json:"api_key"`
}

// RegisterWithMaster attempts to register agent with master
// Returns error if master is unreachable or rejects registration
func (d *Discovery) RegisterWithMaster() error {
	info := AgentInfo{
		NodeID:      d.cfg.Node.ID,
		NodeName:    d.cfg.Node.Name,
		IP: d.cfg.Network.IP,
		Port:        d.cfg.Network.ListenPort,
		APIKey:      d.cfg.Security.APIKey,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal agent info: %w", err)
	}

	url := fmt.Sprintf("%s/api/agents/register", d.masterURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", d.cfg.Security.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("master rejected registration (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("master rejected registration: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully registered with master at %s", d.masterURL)
	d.registered = true
	return nil
}

// UnregisterFromMaster removes agent from master
func (d *Discovery) UnregisterFromMaster() error {
	d.stopHeartbeat <- true

	url := fmt.Sprintf("%s/api/agents/%s", d.masterURL, d.cfg.Node.ID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", d.cfg.Security.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unregister: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Unregistered from master")
	d.registered = false
	return nil
}

// StartRegistrationLoop attempts registration with retry
// This should be called when agent starts
func (d *Discovery) StartRegistrationLoop() {
	go func() {
		for {
			if d.registered {
				// If registered, check if we need to re-register
				if err := d.RegisterWithMaster(); err != nil {
					log.Printf("Re-registration failed: %v", err)
					d.registered = false
				} else {
					// If already registered, just verify connection
					time.Sleep(30 * time.Second)
					continue
				}
			}

			// Try to register
			if err := d.RegisterWithMaster(); err != nil {
				log.Printf("Registration failed: %v, retrying in 5s...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Start heartbeat once registered
			go d.StartHeartbeat()
			break
		}
	}()
}

// StartHeartbeat sends periodic heartbeats to master
// Heartbeats inform master that agent is still alive
// If master doesn't receive heartbeat, agent is marked as offline
func (d *Discovery) StartHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := d.sendHeartbeat(); err != nil {
					log.Printf("Heartbeat failed: %v", err)
					// If heartbeat fails, try to re-register
					d.registered = false
					go d.StartRegistrationLoop()
					return
				}
			case <-d.stopHeartbeat:
				ticker.Stop()
				return
			}
		}
	}()
}

// sendHeartbeat sends a single heartbeat to master
func (d *Discovery) sendHeartbeat() error {
	url := fmt.Sprintf("%s/api/agents/%s/heartbeat", d.masterURL, d.cfg.Node.ID)
	
	info := AgentInfo{
		NodeID:      d.cfg.Node.ID,
		NodeName:    d.cfg.Node.Name,
		IP: d.cfg.Network.IP,
		Port:        d.cfg.Network.ListenPort,
		APIKey:      d.cfg.Security.APIKey,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", d.cfg.Security.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat rejected: %d", resp.StatusCode)
	}

	return nil
}

// IsRegistered returns true if agent is registered with master
func (d *Discovery) IsRegistered() bool {
	return d.registered
}
