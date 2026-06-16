package gpio



// GPIO defines the interface for GPIO operations
type GPIO interface {
	// Init initializes a pin for the given mode ("input" or "output")
	Init(pin uint8, mode string) error

	// Set sets a pin HIGH (true) or LOW (false)
	Set(pin uint8, high bool) error

	// Get reads the current state of a pin
	Get(pin uint8) (bool, error)

	// Toggle flips the state of an output pin
	Toggle(pin uint8) error

	// Close releases a pin
	Close(pin uint8) error

	// Watch registers a callback for edge detection on input pins
	// edge can be "rising", "falling", or "both"
	Watch(pin uint8, edge string, handler func()) error

	// StopWatch removes a watch from a pin
	StopWatch(pin uint8) error
}


