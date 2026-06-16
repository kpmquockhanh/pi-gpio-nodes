package automation

import (
	"pi-gpio-dashboard/config"
	"time"
)

// Rule represents a complete automation rule
type Rule struct {
	ID        string
	Name      string
	Enabled   bool
	Trigger   config.TriggerConfig
	Actions   []config.ActionStepConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}
