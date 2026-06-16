package node

import (
	"fmt"
	"sync"
	"time"

	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/gpio"
)

// PinState represents the current state of a pin
type PinState struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	BCM       int         `json:"bcm"`
	Type      string      `json:"type"`
	Mode      string      `json:"mode"`
	State     interface{} `json:"state"` // bool for digital, float64 for analog
	LastUpdate time.Time  `json:"last_update"`
}

// NodeState represents the full state of a node
type NodeState struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Role   string              `json:"role"`
	Status string              `json:"status"`
	Pins   map[string]*PinState `json:"pins"`
}

// Manager manages local GPIO pins and their state
type Manager struct {
	mu      sync.RWMutex
	config  *config.Config
	gpio    gpio.GPIO
	pins    map[string]*PinState
	handlers map[string]func(string, interface{})
}

// NewManager creates a new node manager
func NewManager(cfg *config.Config, g gpio.GPIO) *Manager {
	return &Manager{
		config:   cfg,
		gpio:     g,
		pins:     make(map[string]*PinState),
		handlers: make(map[string]func(string, interface{})),
	}
}

// Initialize sets up all configured pins
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pinCfg := range m.config.Pins {
		if err := m.gpio.Init(uint8(pinCfg.BCM), pinCfg.Mode); err != nil {
			return fmt.Errorf("failed to init pin %s (BCM %d): %w", pinCfg.ID, pinCfg.BCM, err)
		}

		state := &PinState{
			ID:     pinCfg.ID,
			Name:   pinCfg.Name,
			BCM:    pinCfg.BCM,
			Type:   pinCfg.Type,
			Mode:   pinCfg.Mode,
			State:  false, // default LOW
			LastUpdate: time.Now(),
		}

		// Set default state for outputs
		if pinCfg.Mode == "output" {
			if pinCfg.DefaultState == "high" {
				m.gpio.Set(uint8(pinCfg.BCM), true)
				state.State = true
			} else {
				m.gpio.Set(uint8(pinCfg.BCM), false)
				state.State = false
			}
		} else {
			// Read current input state
			if val, err := m.gpio.Get(uint8(pinCfg.BCM)); err == nil {
				state.State = val
			}
		}

		m.pins[pinCfg.ID] = state
	}

	return nil
}

// GetState returns the current state of a pin
func (m *Manager) GetState(pinID string) (*PinState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.pins[pinID]
	if !ok {
		return nil, fmt.Errorf("pin %s not found", pinID)
	}

	// Refresh state from GPIO
	pinCfg := m.config.GetPinConfig(pinID)
	if pinCfg != nil {
		if val, err := m.gpio.Get(uint8(pinCfg.BCM)); err == nil {
			state.State = val
			state.LastUpdate = time.Now()
		}
	}

	return state, nil
}

// GetAllStates returns all pin states
func (m *Manager) GetAllStates() map[string]*PinState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*PinState)
	for id, state := range m.pins {
		result[id] = state
	}
	return result
}

// ExecuteAction performs an action on a pin
func (m *Manager) ExecuteAction(pinID string, action string, params map[string]interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pinCfg := m.config.GetPinConfig(pinID)
	if pinCfg == nil {
		return nil, fmt.Errorf("pin %s not found in config", pinID)
	}

	state, ok := m.pins[pinID]
	if !ok {
		return nil, fmt.Errorf("pin %s not initialized", pinID)
	}

	var result interface{}
	var err error

	switch action {
	case "set":
		high, ok := params["state"].(bool)
		if !ok {
			return nil, fmt.Errorf("set action requires 'state' parameter")
		}
		err = m.gpio.Set(uint8(pinCfg.BCM), high)
		if err == nil {
			state.State = high
			result = high
		}

	case "toggle":
		err = m.gpio.Toggle(uint8(pinCfg.BCM))
		if err == nil {
			newState, _ := m.gpio.Get(uint8(pinCfg.BCM))
			state.State = newState
			result = newState
		}

	case "pulse":
		durationMs, ok := params["duration_ms"].(float64)
		if !ok {
			durationMsInt, ok := params["duration_ms"].(int)
			if !ok {
				return nil, fmt.Errorf("pulse action requires 'duration_ms' parameter")
			}
			durationMs = float64(durationMsInt)
		}

		// Validate max pulse duration
		if pinCfg.Actions["pulse"].MaxMs != nil && int(durationMs) > *pinCfg.Actions["pulse"].MaxMs {
			return nil, fmt.Errorf("pulse duration %dms exceeds maximum %dms", int(durationMs), *pinCfg.Actions["pulse"].MaxMs)
		}

		// Execute pulse
		err = m.gpio.Set(uint8(pinCfg.BCM), true)
		if err == nil {
			state.State = true
			go func() {
				time.Sleep(time.Duration(durationMs) * time.Millisecond)
				m.gpio.Set(uint8(pinCfg.BCM), false)
				state.State = false
				state.LastUpdate = time.Now()
				m.notifyHandlers(pinID, false)
			}()
			result = true
		}

	case "blink":
		times, _ := params["times"].(float64)
		if times == 0 && pinCfg.Actions["blink"].DefaultTimes != nil {
			times = float64(*pinCfg.Actions["blink"].DefaultTimes)
		}
		intervalMs, _ := params["interval_ms"].(float64)
		if intervalMs == 0 && pinCfg.Actions["blink"].IntervalMs != nil {
			intervalMs = float64(*pinCfg.Actions["blink"].IntervalMs)
		}

		go func() {
			for i := 0; i < int(times); i++ {
				m.gpio.Set(uint8(pinCfg.BCM), true)
				state.State = true
				m.notifyHandlers(pinID, true)
				time.Sleep(time.Duration(intervalMs) * time.Millisecond)
				m.gpio.Set(uint8(pinCfg.BCM), false)
				state.State = false
				m.notifyHandlers(pinID, false)
				time.Sleep(time.Duration(intervalMs) * time.Millisecond)
			}
			state.LastUpdate = time.Now()
		}()
		result = int(times)

	case "read":
		val, err := m.gpio.Get(uint8(pinCfg.BCM))
		if err == nil {
			state.State = val
			result = val
		}

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		return nil, err
	}

	state.LastUpdate = time.Now()
	m.notifyHandlers(pinID, state.State)
	return result, nil
}

// OnStateChange registers a handler for state changes
func (m *Manager) OnStateChange(pinID string, handler func(string, interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[pinID] = handler
}

func (m *Manager) notifyHandlers(pinID string, state interface{}) {
	if handler, ok := m.handlers[pinID]; ok {
		go handler(pinID, state)
	}
}

// GetNodeState returns the full node state
func (m *Manager) GetNodeState() *NodeState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &NodeState{
		ID:     m.config.Node.ID,
		Name:   m.config.Node.Name,
		Role:   m.config.Node.Role,
		Status: "online",
		Pins:   m.pins,
	}
}

// Close cleans up all pins
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pinCfg := range m.config.Pins {
		m.gpio.Close(uint8(pinCfg.BCM))
	}
}
