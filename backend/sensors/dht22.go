package sensors

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pi-gpio-dashboard/events"
	"pi-gpio-dashboard/gpio"
)

// DHT22 represents a DHT22/AM2302 temperature and humidity sensor
type DHT22 struct {
	id         string
	pin        int    // GPIO pin number
	devicePath string // Path to sysfs device or DHT driver
	eventBus   *events.EventBus
	running    bool
	stopChan   chan struct{}
}

// NewDHT22 creates a new DHT22 sensor
func NewDHT22(id string, pin int, eventBus *events.EventBus) *DHT22 {
	return &DHT22{
		id:         id,
		pin:        pin,
		devicePath: fmt.Sprintf("/sys/bus/iio/devices/iio:device0"), // Default path
		eventBus:   eventBus,
		stopChan:   make(chan struct{}),
	}
}

func (d *DHT22) ID() string   { return d.id }
func (d *DHT22) Type() string { return "dht22" }

// Read reads temperature and humidity from the sensor
func (d *DHT22) Read() (Reading, error) {
	// Try to read from sysfs
	temp, humidity, err := d.readSysfs()
	if err != nil {
		return Reading{}, err
	}

	return Reading{
		Value:     temp,
		Unit:      "celsius",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"humidity":    humidity,
			"sensor_type": "dht22",
			"pin":         d.pin,
		},
	}, nil
}

func (d *DHT22) readSysfs() (float64, float64, error) {
	// Read temperature
	tempPath := filepath.Join(d.devicePath, "in_temp_input")
	tempData, err := os.ReadFile(tempPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read temperature: %w", err)
	}

	tempMilliC, err := strconv.Atoi(strings.TrimSpace(string(tempData)))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid temperature data: %w", err)
	}
	temp := float64(tempMilliC) / 1000.0

	// Read humidity
	humPath := filepath.Join(d.devicePath, "in_humidityrelative_input")
	humData, err := os.ReadFile(humPath)
	if err != nil {
		return temp, 0, fmt.Errorf("failed to read humidity: %w", err)
	}

	humMilli, err := strconv.Atoi(strings.TrimSpace(string(humData)))
	if err != nil {
		return temp, 0, fmt.Errorf("invalid humidity data: %w", err)
	}
	humidity := float64(humMilli) / 1000.0

	return temp, humidity, nil
}

// Start begins periodic reading
func (d *DHT22) Start() error {
	if d.running {
		return nil
	}

	d.running = true
	go d.readingLoop()
	return nil
}

// Stop stops periodic reading
func (d *DHT22) Stop() error {
	if !d.running {
		return nil
	}

	close(d.stopChan)
	d.running = false
	return nil
}

func (d *DHT22) readingLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			reading, err := d.Read()
			if err != nil {
				continue
			}

			if d.eventBus != nil {
				d.eventBus.Publish(events.Event{
					Type:      "sensor_reading",
					Source:    d.id,
					Timestamp: reading.Timestamp,
					Data: map[string]interface{}{
						"temperature": reading.Value,
						"unit":        reading.Unit,
						"humidity":    reading.Metadata["humidity"],
						"metadata":    reading.Metadata,
					},
				})
			}
		}
	}
}

// PIR represents a PIR motion sensor
type PIR struct {
	id       string
	pin      int
	gpio     gpio.GPIO
	eventBus *events.EventBus
	running  bool
	stopChan chan struct{}
	lastState bool
}

// NewPIR creates a new PIR motion sensor
func NewPIR(id string, pin int, g gpio.GPIO, eventBus *events.EventBus) *PIR {
	return &PIR{
		id:       id,
		pin:      pin,
		gpio:     g,
		eventBus: eventBus,
		stopChan: make(chan struct{}),
	}
}

func (p *PIR) ID() string   { return p.id }
func (p *PIR) Type() string { return "pir" }

// Read reads the current state of the PIR sensor
func (p *PIR) Read() (Reading, error) {
	// Read from GPIO
	state, err := p.gpio.Get(uint8(p.pin))
	if err != nil {
		return Reading{}, err
	}

	return Reading{
		Value:     boolToFloat(state),
		Unit:      "boolean",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"motion":      state,
			"sensor_type": "pir",
			"pin":         p.pin,
		},
	}, nil
}

func (p *PIR) Start() error {
	if p.running {
		return nil
	}

	p.running = true
	go p.monitoringLoop()
	return nil
}

func (p *PIR) Stop() error {
	if !p.running {
		return nil
	}

	close(p.stopChan)
	p.running = false
	return nil
}

func (p *PIR) monitoringLoop() {
	// Check every 100ms for motion
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			reading, err := p.Read()
			if err != nil {
				continue
			}

			motion := reading.Value > 0
			if motion != p.lastState {
				p.lastState = motion

				if p.eventBus != nil {
					p.eventBus.Publish(events.Event{
						Type:      "motion_detected",
						Source:    p.id,
						Timestamp: reading.Timestamp,
						Data: map[string]interface{}{
							"motion":   motion,
							"pin":      p.pin,
							"metadata": reading.Metadata,
						},
					})
				}
			}
		}
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
