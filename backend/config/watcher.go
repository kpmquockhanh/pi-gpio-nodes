package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// WatchConfig watches a config file for changes and reloads it
func WatchConfig(path string, onChange func(*Config)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Config file modified:", event.Name)
					cfg, err := Load(path)
					if err != nil {
						log.Println("Error reloading config:", err)
						continue
					}
					onChange(cfg)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Config watcher error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		return err
	}
	<-done
	return nil
}

// WatchConfigAsync watches a config file for changes asynchronously
func WatchConfigAsync(path string, onChange func(*Config)) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Config file modified:", event.Name)
					cfg, err := Load(path)
					if err != nil {
						log.Println("Error reloading config:", err)
						continue
					}
					onChange(cfg)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Config watcher error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return watcher, nil
}

// StopWatch stops watching the config file
func StopWatch(watcher *fsnotify.Watcher) {
	if watcher != nil {
		watcher.Close()
	}
}

// IsFileExists checks if a file exists
func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDefaultConfigFile creates a default config file if it doesn't exist
func CreateDefaultConfigFile(path string) error {
	if IsFileExists(path) {
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	defaultConfig := `# Pi GPIO Dashboard Configuration

[node]
id = "master-pi"
name = "Master Pi"
role = "master"

[network]
listen_port = 8080

[security]
api_key = "your-secure-api-key"

[[pin]]
id = "pc-power"
name = "PC Power Relay"
bcm = 17
type = "relay"
mode = "output"
default_state = "low"

  [pin.actions]
  pulse = { default_ms = 500, max_ms = 5000 }
  toggle = {}
  set = {}
`

	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return err
	}

	log.Println("Created default config file:", path)
	return nil
}
