package automation

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/db"
)

// Registry manages automation rules and their state
type Registry struct {
	mu      sync.RWMutex
	rules   map[string]*Rule
	history map[string][]*ExecutionRecord
}

// ExecutionRecord tracks a single rule execution
type ExecutionRecord struct {
	RuleID      string
	TriggeredAt time.Time
	Trigger     config.TriggerConfig
	Actions     []ActionResult
	Success     bool
	Duration    time.Duration
}

// ActionResult tracks a single action execution
type ActionResult struct {
	Step      int
	Type      string
	Node      string
	Pin       string
	Action    string
	Success   bool
	Error     string
	Duration  time.Duration
}

// NewRegistry creates a new automation registry
func NewRegistry() *Registry {
	return &Registry{
		rules:   make(map[string]*Rule),
		history: make(map[string][]*ExecutionRecord),
	}
}

// LoadFromDB loads all automations from database
func (r *Registry) LoadFromDB(database *db.DB) error {
	automations, err := database.GetAutomations()
	if err != nil {
		return fmt.Errorf("failed to load automations: %w", err)
	}

	for _, auto := range automations {
		if err := r.LoadFromAutomation(&auto); err != nil {
			log.Printf("Failed to load automation %s: %v", auto.ID, err)
			continue
		}
	}

	log.Printf("Loaded %d automation rules from database", len(r.rules))
	return nil
}

// LoadFromAutomation loads a single automation from database model
func (r *Registry) LoadFromAutomation(auto *db.Automation) error {
	// Parse the JSON config
	var config struct {
		Trigger config.TriggerConfig   `json:"trigger"`
		Actions []config.ActionStepConfig `json:"actions"`
	}

	if err := json.Unmarshal([]byte(auto.Config), &config); err != nil {
		return fmt.Errorf("failed to parse automation config: %w", err)
	}

	rule := &Rule{
		ID:        auto.ID,
		Name:      auto.Name,
		Enabled:   auto.Enabled,
		Trigger:   config.Trigger,
		Actions:   config.Actions,
		CreatedAt: time.Unix(auto.CreatedAt, 0),
		UpdatedAt: time.Unix(auto.UpdatedAt, 0),
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[rule.ID] = rule

	return nil
}

// SaveToDB saves a rule to database
func (r *Registry) SaveToDB(rule *Rule, database *db.DB) error {
	config := struct {
		Trigger config.TriggerConfig   `json:"trigger"`
		Actions []config.ActionStepConfig `json:"actions"`
	}{
		Trigger: rule.Trigger,
		Actions: rule.Actions,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	auto := &db.Automation{
		ID:      rule.ID,
		Name:    rule.Name,
		Enabled: rule.Enabled,
		Config:  string(configJSON),
	}

	if err := database.SaveAutomation(auto); err != nil {
		return fmt.Errorf("failed to save automation: %w", err)
	}

	return nil
}

// Add adds a rule to the registry
func (r *Registry) Add(rule *Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[rule.ID] = rule
}

// Remove removes a rule from the registry
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules, id)
	delete(r.history, id)
}

// Get returns a rule by ID
func (r *Registry) Get(id string) (*Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[id]
	return rule, ok
}

// GetAll returns all rules
func (r *Registry) GetAll() []*Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]*Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetEnabled returns all enabled rules
func (r *Registry) GetEnabled() []*Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]*Rule, 0)
	for _, rule := range r.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}

// Enable enables/disables a rule
func (r *Registry) Enable(id string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule, ok := r.rules[id]
	if !ok {
		return fmt.Errorf("rule not found: %s", id)
	}

	rule.Enabled = enabled
	rule.UpdatedAt = time.Now()
	return nil
}

// RecordExecution records an execution in history
func (r *Registry) RecordExecution(record *ExecutionRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()

	history := r.history[record.RuleID]
	if len(history) >= 100 {
		// Keep only last 100 records
		history = history[1:]
	}
	history = append(history, record)
	r.history[record.RuleID] = history
}

// GetHistory returns execution history for a rule
func (r *Registry) GetHistory(ruleID string) []*ExecutionRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.history[ruleID]
}

// GetAllHistory returns all execution history
func (r *Registry) GetAllHistory() map[string][]*ExecutionRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*ExecutionRecord)
	for k, v := range r.history {
		result[k] = v
	}
	return result
}

// Count returns total number of rules
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rules)
}

// CountEnabled returns number of enabled rules
func (r *Registry) CountEnabled() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, rule := range r.rules {
		if rule.Enabled {
			count++
		}
	}
	return count
}

// CreateRule creates a new rule from user input
func CreateRule(name string, trigger config.TriggerConfig, actions []config.ActionStepConfig) *Rule {
	return &Rule{
		ID:        generateRuleID(),
		Name:      name,
		Enabled:   true,
		Trigger:   trigger,
		Actions:   actions,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func generateRuleID() string {
	return fmt.Sprintf("auto-%d", time.Now().UnixNano())
}
