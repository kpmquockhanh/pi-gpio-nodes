package automation

import (
	"fmt"
	"time"

	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/events"
)

// TriggerEvaluator evaluates trigger conditions
type TriggerEvaluator struct {
	// Track previous states for edge detection
	pinHistory map[string]*pinStateRecord
}

type pinStateRecord struct {
	State     interface{}
	Timestamp time.Time
}

// NewTriggerEvaluator creates a new trigger evaluator
func NewTriggerEvaluator() *TriggerEvaluator {
	return &TriggerEvaluator{
		pinHistory: make(map[string]*pinStateRecord),
	}
}

// Evaluate checks if a rule's trigger matches an event
func (te *TriggerEvaluator) Evaluate(rule *Rule, event events.Event) bool {
	trigger := rule.Trigger

	// Check node match (if specified)
	if trigger.Node != "" && trigger.Node != event.Source {
		return false
	}

	// Check pin match
	if trigger.Pin != "" && trigger.Pin != event.PinID {
		return false
	}

	switch trigger.Type {
	case "pin_state":
		return te.evaluatePinState(trigger, event)
	case "value_threshold":
		return te.evaluateValueThreshold(trigger, event)
	case "timer":
		return te.evaluateTimer(trigger, event)
	case "long_press":
		return te.evaluateLongPress(trigger, event)
	default:
		return false
	}
}

func (te *TriggerEvaluator) evaluatePinState(trigger config.TriggerConfig, event events.Event) bool {
	if event.Type != "pin_state_change" {
		return false
	}

	state, ok := event.Data["state"]
	if !ok {
		return false
	}

	// Update history
	key := fmt.Sprintf("%s/%s", event.Source, event.PinID)
	previous := te.pinHistory[key]
	current := &pinStateRecord{State: state, Timestamp: event.Timestamp}
	defer func() { te.pinHistory[key] = current }()

	switch trigger.Condition {
	case "HIGH", "ON":
		return state == true || state == "HIGH" || state == "ON"
	case "LOW", "OFF":
		return state == false || state == "LOW" || state == "OFF"
	case "rising_edge":
		if previous == nil {
			return false
		}
		prevState := isHigh(previous.State)
		currState := isHigh(state)
		return !prevState && currState
	case "falling_edge":
		if previous == nil {
			return false
		}
		prevState := isHigh(previous.State)
		currState := isHigh(state)
		return prevState && !currState
	case "change":
		if previous == nil {
			return false
		}
		return previous.State != state
	default:
		return false
	}
}

func (te *TriggerEvaluator) evaluateValueThreshold(trigger config.TriggerConfig, event events.Event) bool {
	if event.Type != "value_threshold" {
		return false
	}

	value, ok := event.Data["value"].(float64)
	if !ok {
		return false
	}

	// Update history
	key := fmt.Sprintf("%s/%s", event.Source, event.PinID)
	previous := te.pinHistory[key]
	current := &pinStateRecord{State: value, Timestamp: event.Timestamp}
	defer func() { te.pinHistory[key] = current }()

	switch trigger.Condition {
	case "greater_than":
		return value > trigger.Threshold
	case "less_than":
		return value < trigger.Threshold
	case "equals":
		return value == trigger.Threshold
	case "greater_than_or_equal":
		return value >= trigger.Threshold
	case "less_than_or_equal":
		return value <= trigger.Threshold
	case "crosses_above":
		if previous == nil {
			return false
		}
		prevValue, ok := previous.State.(float64)
		if !ok {
			return false
		}
		return prevValue <= trigger.Threshold && value > trigger.Threshold
	case "crosses_below":
		if previous == nil {
			return false
		}
		prevValue, ok := previous.State.(float64)
		if !ok {
			return false
		}
		return prevValue >= trigger.Threshold && value < trigger.Threshold
	default:
		return false
	}
}

func (te *TriggerEvaluator) evaluateTimer(trigger config.TriggerConfig, event events.Event) bool {
	if event.Type != "timer" {
		return false
	}

	// Check if this is the right timer
	timerID, _ := event.Data["timer_id"].(string)
	if trigger.Interval != "" && trigger.Interval != timerID {
		return false
	}

	return true
}

func (te *TriggerEvaluator) evaluateLongPress(trigger config.TriggerConfig, event events.Event) bool {
	if event.Type != "pin_state_change" {
		return false
	}

	state, ok := event.Data["state"]
	if !ok {
		return false
	}

	// Check if state is HIGH
	if !isHigh(state) {
		return false
	}

	// Check if duration exceeds threshold
	key := fmt.Sprintf("%s/%s", event.Source, event.PinID)
	previous := te.pinHistory[key]
	
	if previous == nil {
		// First time seeing HIGH, record it
		te.pinHistory[key] = &pinStateRecord{State: state, Timestamp: event.Timestamp}
		return false
	}

	// Check if we've been holding long enough
	duration := event.Timestamp.Sub(previous.Timestamp)
	requiredDuration := time.Duration(trigger.DurationMs) * time.Millisecond

	if duration >= requiredDuration {
		// Reset the history to prevent repeated triggering
		te.pinHistory[key] = nil
		return true
	}

	return false
}

func isHigh(state interface{}) bool {
	switch v := state.(type) {
	case bool:
		return v
	case string:
		return v == "HIGH" || v == "ON"
	case float64:
		return v > 0
	case int:
		return v > 0
	default:
		return false
	}
}

// Cleanup removes old history entries
func (te *TriggerEvaluator) Cleanup(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	for key, record := range te.pinHistory {
		if record != nil && record.Timestamp.Before(cutoff) {
			delete(te.pinHistory, key)
		}
	}
}
