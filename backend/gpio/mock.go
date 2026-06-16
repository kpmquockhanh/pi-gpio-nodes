package gpio

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// MockGPIO is a mock implementation for development/testing
type MockGPIO struct {
	mu       sync.RWMutex
	states   map[uint8]bool
	watches  map[uint8]*watchConfig
}

type watchConfig struct {
	edge    string
	handler func()
}

// NewMockGPIO creates a mock GPIO implementation
func NewMockGPIO() (*MockGPIO, error) {
	log.Println("Using MOCK GPIO - no real hardware access")
	return &MockGPIO{
		states:  make(map[uint8]bool),
		watches: make(map[uint8]*watchConfig),
	}, nil
}

func (m *MockGPIO) Init(pin uint8, mode string) error {
	if mode != "input" && mode != "output" {
		return fmt.Errorf("invalid mode: %s", mode)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[pin] = false
	return nil
}

func (m *MockGPIO) Set(pin uint8, high bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[pin] = high
	
	// Check if there's a watch and trigger it
	if watch, ok := m.watches[pin]; ok {
		go m.checkWatch(pin, watch)
	}
	return nil
}

func (m *MockGPIO) Get(pin uint8) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// For input pins, simulate random state changes
	if _, ok := m.watches[pin]; ok {
		// Simulate occasional state changes
		if rand.Float32() < 0.1 {
			return !m.states[pin], nil
		}
	}
	
	return m.states[pin], nil
}

func (m *MockGPIO) Toggle(pin uint8) error {
	current, err := m.Get(pin)
	if err != nil {
		return err
	}
	return m.Set(pin, !current)
}

func (m *MockGPIO) Close(pin uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, pin)
	delete(m.watches, pin)
	return nil
}

func (m *MockGPIO) Watch(pin uint8, edge string, handler func()) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.watches[pin] = &watchConfig{edge: edge, handler: handler}
	
	// Start a background goroutine to simulate edge detection
	go m.simulateEdgeDetection(pin, edge, handler)
	return nil
}

func (m *MockGPIO) StopWatch(pin uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.watches, pin)
	return nil
}

func (m *MockGPIO) checkWatch(pin uint8, watch *watchConfig) {
	// Simplified - just trigger on any change for simulation
	if watch.edge != "" && watch.handler != nil {
		watch.handler()
	}
}

func (m *MockGPIO) simulateEdgeDetection(pin uint8, edge string, handler func()) {
	lastVal, _ := m.Get(pin)
	for {
		time.Sleep(100 * time.Millisecond)
		
		m.mu.RLock()
		_, stillWatching := m.watches[pin]
		m.mu.RUnlock()
		
		if !stillWatching {
			return
		}
		
		val, _ := m.Get(pin)
		if val != lastVal {
			triggered := false
			switch edge {
			case "rising":
				triggered = !lastVal && val
			case "falling":
				triggered = lastVal && !val
			case "both":
				triggered = true
			}
			if triggered {
				handler()
			}
			lastVal = val
		}
	}
}

// CloseAll closes all pins
func (m *MockGPIO) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states = make(map[uint8]bool)
	m.watches = make(map[uint8]*watchConfig)
}

// New creates the appropriate GPIO implementation based on environment
func New() (GPIO, error) {
	if os.Getenv("MOCK_GPIO") == "true" {
		return NewMockGPIO()
	}
	return NewSysfsGPIO()
}
