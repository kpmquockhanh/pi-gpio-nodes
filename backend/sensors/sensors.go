package sensors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pi-gpio-dashboard/events"
	"pi-gpio-dashboard/gpio"
)

// Sensor defines the interface for all sensors
type Sensor interface {
	ID() string
	Type() string
	Read() (Reading, error)
	Start() error
	Stop() error
}

// Reading represents a sensor reading
type Reading struct {
	Value     float64
	Unit      string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// SensorManager manages all sensors for a node
type SensorManager struct {
	mu       sync.RWMutex
	sensors  map[string]Sensor
	eventBus *events.EventBus
	gpio     gpio.GPIO
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSensorManager creates a new sensor manager
func NewSensorManager(eventBus *events.EventBus, g gpio.GPIO) *SensorManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SensorManager{
		sensors:  make(map[string]Sensor),
		eventBus: eventBus,
		gpio:     g,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AddSensor adds a sensor to the manager
func (sm *SensorManager) AddSensor(sensor Sensor) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sensors[sensor.ID()]; exists {
		return fmt.Errorf("sensor %s already exists", sensor.ID())
	}

	sm.sensors[sensor.ID()] = sensor
	
	if sm.running {
		return sensor.Start()
	}
	
	return nil
}

// RemoveSensor removes a sensor from the manager
func (sm *SensorManager) RemoveSensor(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sensor, exists := sm.sensors[id]
	if !exists {
		return fmt.Errorf("sensor %s not found", id)
	}

	if sm.running {
		sensor.Stop()
	}

	delete(sm.sensors, id)
	return nil
}

// GetSensor returns a sensor by ID
func (sm *SensorManager) GetSensor(id string) (Sensor, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sensor, exists := sm.sensors[id]
	return sensor, exists
}

// GetAllSensors returns all sensors
func (sm *SensorManager) GetAllSensors() map[string]Sensor {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]Sensor)
	for id, sensor := range sm.sensors {
		result[id] = sensor
	}
	return result
}

// Start starts all sensors
func (sm *SensorManager) Start() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return
	}

	sm.running = true
	for _, sensor := range sm.sensors {
		go sensor.Start()
	}

	// Start periodic collection
	go sm.collectionLoop()
}

// Stop stops all sensors
func (sm *SensorManager) Stop() {
	sm.cancel()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.running = false
	for _, sensor := range sm.sensors {
		sensor.Stop()
	}
}

func (sm *SensorManager) collectionLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.collectReadings()
		}
	}
}

func (sm *SensorManager) collectReadings() {
	sm.mu.RLock()
	sensors := make(map[string]Sensor)
	for id, sensor := range sm.sensors {
		sensors[id] = sensor
	}
	sm.mu.RUnlock()

	for id, sensor := range sensors {
		reading, err := sensor.Read()
		if err != nil {
			continue
		}

		// Publish event
		if sm.eventBus != nil {
			sm.eventBus.Publish(events.Event{
				Type:      "sensor_reading",
				Source:    id,
				Timestamp: reading.Timestamp,
				Data: map[string]interface{}{
					"value":    reading.Value,
					"unit":     reading.Unit,
					"metadata": reading.Metadata,
				},
			})
		}
	}
}


