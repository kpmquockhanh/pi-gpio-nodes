package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the full TOML configuration for a Pi node
type Config struct {
	Node        NodeConfig        `toml:"node"`
	Network     NetworkConfig     `toml:"network"`
	Security    SecurityConfig    `toml:"security"`
	Pins        []PinConfig       `toml:"pins"`
	Automations []AutomationConfig `toml:"automations"`
}

// NodeConfig defines this Pi's identity
type NodeConfig struct {
	ID   string `toml:"id"`
	Name string `toml:"name"`
	Role string `toml:"role"` // "master" or "agent"
}

// NetworkConfig defines network settings
type NetworkConfig struct {
	ListenPort  int    `toml:"listen_port"`
	IP string `toml:"ip"`
	MasterNode  string `toml:"master_node"` // Agents only
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	APIKey string `toml:"api_key"`
}

// PinConfig defines a GPIO pin configuration
type PinConfig struct {
	ID           string                 `toml:"id"`
	Name         string                 `toml:"name"`
	BCM          int                    `toml:"bcm"`
	Type         string                 `toml:"type"`         // "relay", "led", "button", "dht22"
	Mode         string                 `toml:"mode"`         // "output", "input"
	DefaultState string                 `toml:"default_state"` // "low", "high"
	Pull         string                 `toml:"pull"`          // "up", "down", "none"
	DebounceMs   int                    `toml:"debounce_ms"`
	PollInterval int                    `toml:"poll_interval_sec"`
	Actions      map[string]ActionConfig `toml:"actions"`
}

// ActionConfig defines available actions for a pin
type ActionConfig struct {
	DefaultMs *int `toml:"default_ms,omitempty"`
	MaxMs     *int `toml:"max_ms,omitempty"`
	MaxTimes  *int `toml:"max_times,omitempty"`
	DefaultTimes *int `toml:"default_times,omitempty"`
	IntervalMs   *int `toml:"interval_ms,omitempty"`
}

// AutomationConfig defines an automation rule
type AutomationConfig struct {
	ID      string          `toml:"id"`
	Name    string          `toml:"name"`
	Enabled bool            `toml:"enabled"`
	Trigger TriggerConfig   `toml:"trigger"`
	Actions []ActionStepConfig `toml:"automation.action"`
}

// TriggerConfig defines what triggers an automation
type TriggerConfig struct {
	Type      string  `toml:"type"`       // "pin_state", "value_threshold", "timer"
	Node      string  `toml:"node"`
	Pin       string  `toml:"pin"`
	Condition string  `toml:"condition"`  // "HIGH", "LOW", "rising_edge", "falling_edge", "greater_than", "less_than", "long_press"
	Threshold float64 `toml:"threshold,omitempty"`
	DurationMs int    `toml:"duration_ms,omitempty"`
	Interval  string  `toml:"interval,omitempty"` // for timer triggers
}

// ActionStepConfig defines a single step in an automation action chain
type ActionStepConfig struct {
	Type   string                 `toml:"type"`   // "pin_action", "delay", "notify"
	Node   string                 `toml:"node,omitempty"`
	Pin    string                 `toml:"pin,omitempty"`
	Action string                 `toml:"action,omitempty"`
	Params map[string]interface{} `toml:"params"`
}

// Load reads and parses the TOML configuration file
func Load(path string) (*Config, error) {
	if path == "" {
		path = "/etc/pi-gpio/config.toml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Network.ListenPort == 0 {
		cfg.Network.ListenPort = 8080
	}
	if cfg.Node.Role == "" {
		cfg.Node.Role = "master"
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.Node.ID == "" {
		return fmt.Errorf("node.id is required")
	}
	if c.Node.Name == "" {
		return fmt.Errorf("node.name is required")
	}
	if c.Node.Role != "master" && c.Node.Role != "agent" {
		return fmt.Errorf("node.role must be 'master' or 'agent'")
	}
	if c.Node.Role == "agent" && c.Network.MasterNode == "" {
		return fmt.Errorf("network.master_node is required for agent nodes")
	}
	if c.Node.Role == "agent" && c.Network.IP == "" {
		return fmt.Errorf("network.ip is required for agent nodes")
	}

	// Validate pins
	pinIDs := make(map[string]bool)
	bcmPins := make(map[int]bool)
	for i, pin := range c.Pins {
		if pin.ID == "" {
			return fmt.Errorf("pin[%d].id is required", i)
		}
		if pinIDs[pin.ID] {
			return fmt.Errorf("duplicate pin id: %s", pin.ID)
		}
		pinIDs[pin.ID] = true

		if pin.BCM < 0 || pin.BCM > 27 {
			return fmt.Errorf("pin[%d].bcm must be between 0 and 27", i)
		}
		if bcmPins[pin.BCM] {
			return fmt.Errorf("duplicate bcm pin: %d", pin.BCM)
		}
		bcmPins[pin.BCM] = true
	}

	return nil
}

// IsMaster returns true if this node is configured as master
func (c *Config) IsMaster() bool {
	return c.Node.Role == "master"
}

// GetPinConfig finds a pin configuration by ID
func (c *Config) GetPinConfig(id string) *PinConfig {
	for i := range c.Pins {
		if c.Pins[i].ID == id {
			return &c.Pins[i]
		}
	}
	return nil
}

// Save writes the configuration to a TOML file
func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
