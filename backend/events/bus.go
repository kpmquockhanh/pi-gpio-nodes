package events

import (
	"sync"
	"time"
)

// Event represents a system event
type Event struct {
	Type      string
	Source    string      // Node ID
	PinID     string
	Timestamp time.Time
	Data      map[string]interface{}
}

// EventBus provides pub/sub functionality
type EventBus struct {
	mu        sync.RWMutex
	subscribers map[string][]chan Event
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

// Subscribe registers a channel for specific event types
func (eb *EventBus) Subscribe(eventTypes []string) chan Event {
	ch := make(chan Event, 100)
	
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	for _, eventType := range eventTypes {
		eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	}
	
	return ch
}

// Unsubscribe removes a channel from all event types
func (eb *EventBus) Unsubscribe(ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	for eventType, subs := range eb.subscribers {
		newSubs := make([]chan Event, 0)
		for _, sub := range subs {
			if sub != ch {
				newSubs = append(newSubs, sub)
			}
		}
		eb.subscribers[eventType] = newSubs
	}
	
	close(ch)
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	subs := eb.subscribers[event.Type]
	for _, ch := range subs {
		select {
		case ch <- event:
			// Successfully sent
		default:
			// Channel full, drop event
		}
	}
}

// PublishPinStateChange convenience method
func (eb *EventBus) PublishPinStateChange(nodeID, pinID string, state interface{}) {
	eb.Publish(Event{
		Type:      "pin_state_change",
		Source:    nodeID,
		PinID:     pinID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"state": state,
		},
	})
}

// PublishValueThreshold convenience method
func (eb *EventBus) PublishValueThreshold(nodeID, pinID string, value float64, threshold float64, condition string) {
	eb.Publish(Event{
		Type:      "value_threshold",
		Source:    nodeID,
		PinID:     pinID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"value":     value,
			"threshold": threshold,
			"condition": condition,
		},
	})
}

// PublishTimer convenience method
func (eb *EventBus) PublishTimer(timerID string) {
	eb.Publish(Event{
		Type:      "timer",
		Source:    "system",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"timer_id": timerID,
		},
	})
}