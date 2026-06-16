package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/db"
	"pi-gpio-dashboard/events"
	"pi-gpio-dashboard/master"
	"pi-gpio-dashboard/node"
)

// Engine evaluates automation rules and executes actions
type Engine struct {
	mu          sync.RWMutex
	registry    *Registry
	evaluator   *TriggerEvaluator
	eventBus    *events.EventBus
	manager     *node.Manager
	agentPool   *master.AgentPool
	db          *db.DB
	nodeID      string
	ctx         context.Context
	cancel      context.CancelFunc
	runningRules map[string]bool // Track currently running rules
}

// NewEngine creates a new automation engine
func NewEngine(eventBus *events.EventBus, manager *node.Manager, agentPool *master.AgentPool, database *db.DB, nodeID string) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		registry:     NewRegistry(),
		evaluator:    NewTriggerEvaluator(),
		eventBus:     eventBus,
		manager:      manager,
		agentPool:    agentPool,
		db:           database,
		nodeID:       nodeID,
		ctx:          ctx,
		cancel:       cancel,
		runningRules: make(map[string]bool),
	}
}

// Start begins the automation engine
func (e *Engine) Start() {
	log.Println("Starting automation engine...")

	// Load rules from database
	if err := e.registry.LoadFromDB(e.db); err != nil {
		log.Printf("Failed to load rules: %v", err)
	}

	// Subscribe to events
	ch := e.eventBus.Subscribe([]string{
		"pin_state_change",
		"value_threshold",
		"timer",
	})

	// Start event processing loop
	go e.eventLoop(ch)

	// Start timer-based triggers
	go e.timerLoop()

	// Start history cleanup
	go e.cleanupLoop()
}

// Stop stops the automation engine
func (e *Engine) Stop() {
	log.Println("Stopping automation engine...")
	e.cancel()
}

// AddRule adds a new automation rule
func (e *Engine) AddRule(rule *Rule) error {
	if err := e.registry.SaveToDB(rule, e.db); err != nil {
		return err
	}
	
	e.registry.Add(rule)
	log.Printf("Added automation rule: %s (%s)", rule.Name, rule.ID)
	return nil
}

// UpdateRule updates an existing rule
func (e *Engine) UpdateRule(rule *Rule) error {
	if err := e.registry.SaveToDB(rule, e.db); err != nil {
		return err
	}

	e.registry.Add(rule)
	log.Printf("Updated automation rule: %s (%s)", rule.Name, rule.ID)
	return nil
}

// RemoveRule removes an automation rule
func (e *Engine) RemoveRule(id string) error {
	if err := e.db.DeleteAutomation(id); err != nil {
		return err
	}

	e.registry.Remove(id)
	log.Printf("Removed automation rule: %s", id)
	return nil
}

// GetRule returns a rule by ID
func (e *Engine) GetRule(id string) (*Rule, bool) {
	return e.registry.Get(id)
}

// GetAllRules returns all rules
func (e *Engine) GetAllRules() []*Rule {
	return e.registry.GetAll()
}

// EnableRule enables/disables a rule
func (e *Engine) EnableRule(id string, enabled bool) error {
	if err := e.registry.Enable(id, enabled); err != nil {
		return err
	}

	// Save to database
	rule, ok := e.registry.Get(id)
	if ok {
		return e.registry.SaveToDB(rule, e.db)
	}

	return nil
}

func (e *Engine) eventLoop(ch chan events.Event) {
	for {
		select {
		case <-e.ctx.Done():
			return
		case event := <-ch:
			e.processEvent(event)
		}
	}
}

func (e *Engine) processEvent(event events.Event) {
	rules := e.registry.GetEnabled()

	for _, rule := range rules {
		if e.evaluator.Evaluate(rule, event) {
			// Check if rule is already running (prevent concurrent execution)
			e.mu.Lock()
			if e.runningRules[rule.ID] {
				log.Printf("Rule %s already running, skipping", rule.ID)
				e.mu.Unlock()
				continue
			}
			e.runningRules[rule.ID] = true
			e.mu.Unlock()

			// Execute in goroutine
			go func(r *Rule) {
				e.executeRule(r)
				e.mu.Lock()
				delete(e.runningRules, r.ID)
				e.mu.Unlock()
			}(rule)
		}
	}
}

func (e *Engine) executeRule(rule *Rule) {
	start := time.Now()
	log.Printf("Executing automation rule: %s (%s)", rule.Name, rule.ID)

	record := &ExecutionRecord{
		RuleID:      rule.ID,
		TriggeredAt: start,
		Trigger:     rule.Trigger,
		Actions:     make([]ActionResult, 0, len(rule.Actions)),
		Success:     true,
	}

	for i, action := range rule.Actions {
		stepStart := time.Now()
		log.Printf("  Step %d: %s on %s/%s", i+1, action.Action, action.Node, action.Pin)

		err := e.executeAction(action, rule.ID)
		stepDuration := time.Since(stepStart)

		result := ActionResult{
			Step:     i + 1,
			Type:     action.Type,
			Node:     action.Node,
			Pin:      action.Pin,
			Action:   action.Action,
			Success:  err == nil,
			Duration: stepDuration,
		}

		if err != nil {
			result.Error = err.Error()
			record.Success = false
			log.Printf("  Step %d failed: %v", i+1, err)
		}

		record.Actions = append(record.Actions, result)

		// Handle delay action
		if action.Type == "delay" {
			if ms, ok := action.Params["ms"].(float64); ok {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		}
	}

	record.Duration = time.Since(start)
	e.registry.RecordExecution(record)

	log.Printf("Completed automation rule: %s (success=%v, duration=%v)",
		rule.Name, record.Success, record.Duration)
}

func (e *Engine) executeAction(action config.ActionStepConfig, ruleID string) error {
	targetNode := action.Node
	if targetNode == "" {
		targetNode = e.nodeID
	}

	if action.Type == "delay" {
		return nil
	}

	if action.Type == "pin_action" {
		if targetNode == e.nodeID {
			// Local execution
			result, err := e.manager.ExecuteAction(action.Pin, action.Action, action.Params)
			if err != nil {
				return err
			}
			
			// Log the action
			paramsJSON, _ := json.Marshal(action.Params)
			resultStr, _ := json.Marshal(result)
			e.db.LogAction(targetNode, action.Pin, action.Action, string(paramsJSON), string(resultStr), "automation:"+ruleID)
			return nil
		} else if e.agentPool != nil {
			// Forward to agent
			if err := e.agentPool.ForwardAction(targetNode, action.Pin, action.Action, action.Params); err != nil {
				return err
			}
			
			// Log the action
			paramsJSON, _ := json.Marshal(action.Params)
			e.db.LogAction(targetNode, action.Pin, action.Action, string(paramsJSON), "forwarded", "automation:"+ruleID)
			return nil
		}
	}

	if action.Type == "notify" {
		// Log notification as action
		message, _ := action.Params["message"].(string)
		log.Printf("  [Notification] %s", message)
		e.db.LogAction("system", "", "notify", fmt.Sprintf("%q", message), "sent", "automation:"+ruleID)
		return nil
	}

	return fmt.Errorf("unknown action type: %s", action.Type)
}

func (e *Engine) timerLoop() {
	// Check every minute for timer-based triggers
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			// Check for timer-based triggers
			// This is where you'd implement cron-like scheduling
			// For now, just publish timer events
			
			// Check each rule for interval triggers
			rules := e.registry.GetEnabled()
			for _, rule := range rules {
				if rule.Trigger.Type == "timer" && rule.Trigger.Interval != "" {
					// Parse interval (e.g., "1m", "5m", "1h")
					// For now, trigger on every timer tick
					e.eventBus.Publish(events.Event{
						Type:      "timer",
						Source:    "system",
						Timestamp: time.Now(),
						Data: map[string]interface{}{
							"timer_id": rule.Trigger.Interval,
						},
					})
				}
			}
		}
	}
}

func (e *Engine) cleanupLoop() {
	// Clean up old history every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.evaluator.Cleanup(24 * time.Hour)
		}
	}
}
