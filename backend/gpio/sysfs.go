package gpio

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const sysfsGPIOPath = "/sys/class/gpio"

// ErrGPIOUnavailable is returned when sysfs GPIO is not available
var ErrGPIOUnavailable = fmt.Errorf("sysfs GPIO not available")

// SysfsGPIO implements GPIO using Linux sysfs interface
// Works on all Raspberry Pi models and most Linux boards
// Requires root access or gpio group membership
type SysfsGPIO struct {
	mu       sync.RWMutex
	exported map[uint8]bool
	base     int // gpiochip base offset (e.g., 512 on newer kernels)
}

// detectGPIOBase finds the gpiochip base offset.
// On newer kernels (e.g. Pi OS Bookworm), the base is 512, so BCM 22 -> global 534.
func detectGPIOBase() (int, error) {
	entries, err := os.ReadDir(sysfsGPIOPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read %s: %w", sysfsGPIOPath, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "gpiochip") {
			basePath := filepath.Join(sysfsGPIOPath, name, "base")
			data, err := os.ReadFile(basePath)
			if err != nil {
				continue
			}
			base, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				continue
			}
			// On Raspberry Pi, the main GPIO header is the first (and usually only) gpiochip
			return base, nil
		}
	}

	// Fallback: if no gpiochip found, assume base 0 (older kernels)
	return 0, nil
}

// globalPin converts a BCM pin number to the global GPIO number used by sysfs
func (s *SysfsGPIO) globalPin(pin uint8) int {
	return s.base + int(pin)
}

// pinPath returns the sysfs path for a given BCM pin
func (s *SysfsGPIO) pinPath(pin uint8) string {
	return filepath.Join(sysfsGPIOPath, fmt.Sprintf("gpio%d", s.globalPin(pin)))
}

// NewSysfsGPIO creates a new sysfs GPIO implementation
func NewSysfsGPIO() (*SysfsGPIO, error) {
	// Check if sysfs is available
	if _, err := os.Stat(sysfsGPIOPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sysfs GPIO not available at %s. Ensure GPIO subsystem is enabled and you have proper permissions", sysfsGPIOPath)
	}

	base, err := detectGPIOBase()
	if err != nil {
		return nil, fmt.Errorf("failed to detect GPIO base: %w", err)
	}

	return &SysfsGPIO{
		exported: make(map[uint8]bool),
		base:     base,
	}, nil
}

func (s *SysfsGPIO) Init(pin uint8, mode string) error {
	if mode != "input" && mode != "output" {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	global := s.globalPin(pin)
	pinPath := s.pinPath(pin)

	// Check if pin is already exported
	if _, err := os.Stat(pinPath); os.IsNotExist(err) {
		// Export the pin using the global GPIO number
		exportPath := filepath.Join(sysfsGPIOPath, "export")
		if err := os.WriteFile(exportPath, []byte(strconv.Itoa(global)), 0644); err != nil {
			// Pin might already be exported, which is fine
			errStr := err.Error()
			if !strings.Contains(errStr, "Device or resource busy") && !strings.Contains(errStr, "invalid argument") {
				return fmt.Errorf("failed to export pin %d (global %d): %w", pin, global, err)
			}
		}

		// Wait for pin directory to appear
		for i := 0; i < 10; i++ {
			if _, err := os.Stat(pinPath); err == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Set direction
	directionPath := filepath.Join(pinPath, "direction")
	direction := "in"
	if mode == "output" {
		direction = "out"
	}

	if err := os.WriteFile(directionPath, []byte(direction), 0644); err != nil {
		return fmt.Errorf("failed to set direction for pin %d (global %d): %w", pin, global, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.exported[pin] = true

	return nil
}

func (s *SysfsGPIO) Set(pin uint8, high bool) error {
	value := "0"
	if high {
		value = "1"
	}

	valuePath := filepath.Join(s.pinPath(pin), "value")
	if err := os.WriteFile(valuePath, []byte(value), 0644); err != nil {
		return fmt.Errorf("failed to set pin %d: %w", pin, err)
	}

	return nil
}

func (s *SysfsGPIO) Get(pin uint8) (bool, error) {
	valuePath := filepath.Join(s.pinPath(pin), "value")
	data, err := os.ReadFile(valuePath)
	if err != nil {
		return false, fmt.Errorf("failed to read pin %d: %w", pin, err)
	}

	val := strings.TrimSpace(string(data))
	return val == "1", nil
}

func (s *SysfsGPIO) Toggle(pin uint8) error {
	current, err := s.Get(pin)
	if err != nil {
		return err
	}
	return s.Set(pin, !current)
}

func (s *SysfsGPIO) Close(pin uint8) error {
	global := s.globalPin(pin)
	unexportPath := filepath.Join(sysfsGPIOPath, "unexport")
	if err := os.WriteFile(unexportPath, []byte(strconv.Itoa(global)), 0644); err != nil {
		return fmt.Errorf("failed to unexport pin %d (global %d): %w", pin, global, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.exported, pin)

	return nil
}

func (s *SysfsGPIO) Watch(pin uint8, edge string, handler func()) error {
	// sysfs watch is limited; we can set edge detection
	edgePath := filepath.Join(s.pinPath(pin), "edge")
	edgeValue := "none"
	switch edge {
	case "rising":
		edgeValue = "rising"
	case "falling":
		edgeValue = "falling"
	case "both":
		edgeValue = "both"
	default:
		return fmt.Errorf("invalid edge: %s", edge)
	}

	if err := os.WriteFile(edgePath, []byte(edgeValue), 0644); err != nil {
		return fmt.Errorf("failed to set edge detection for pin %d: %w", pin, err)
	}

	// For actual polling, you'd need to poll the value file or use epoll
	// This is a simplified implementation
	go func() {
		lastVal, _ := s.Get(pin)
		for {
			time.Sleep(10 * time.Millisecond)
			val, err := s.Get(pin)
			if err != nil {
				continue
			}
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
	}()

	return nil
}

func (s *SysfsGPIO) StopWatch(pin uint8) error {
	edgePath := filepath.Join(s.pinPath(pin), "edge")
	return os.WriteFile(edgePath, []byte("none"), 0644)
}

// CloseAll unexports all pins
func (s *SysfsGPIO) CloseAll() {
	s.mu.Lock()
	pins := make([]uint8, 0, len(s.exported))
	for pin := range s.exported {
		pins = append(pins, pin)
	}
	s.mu.Unlock()

	for _, pin := range pins {
		s.Close(pin)
	}
}
