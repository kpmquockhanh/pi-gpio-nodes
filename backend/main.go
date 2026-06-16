package main

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pi-gpio-dashboard/agent"
	"pi-gpio-dashboard/api"
	"pi-gpio-dashboard/automation"
	"pi-gpio-dashboard/config"
	"pi-gpio-dashboard/db"
	"pi-gpio-dashboard/events"
	"pi-gpio-dashboard/gpio"
	"pi-gpio-dashboard/master"
	"pi-gpio-dashboard/node"
	"pi-gpio-dashboard/sensors"
)

//go:embed all:static
var staticFS embed.FS

func main() {
	// Get config path from environment
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/pi-gpio/config.toml"
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, `
================================================================================
  Configuration File Not Found
================================================================================

  Expected: %s

  Please create a configuration file first.

  Example:
    sudo mkdir -p /etc/pi-gpio
    sudo nano %s

  Or set a custom path:
    CONFIG_PATH=/path/to/config.toml go run main.go

================================================================================
`, configPath, configPath)
			os.Exit(1)
		} else {
			log.Fatalf("Failed to load config: %v", err)
		}
	}

	log.Printf("Starting Pi GPIO Dashboard - Node: %s (%s)", cfg.Node.Name, cfg.Node.Role)

	// Initialize GPIO
	g, err := gpio.New()
	if err != nil {
		log.Fatalf("Failed to initialize GPIO: %v", err)
	}

	// Initialize node manager
	manager := node.NewManager(cfg, g)
	if err := manager.Initialize(); err != nil {
		log.Fatalf("Failed to initialize node manager: %v", err)
	}
	defer manager.Close()

	// Setup state change notification for agent mode
	setupStateNotifications(manager, cfg)

	// Start based on role
	if cfg.IsMaster() {
		runMaster(cfg, manager, g)
	} else {
		runAgent(cfg, manager)
	}
}

func runMaster(cfg *config.Config, manager *node.Manager, g gpio.GPIO) {
	// Initialize database (master only)
	dbPath := fmt.Sprintf("pi-gpio-%s.db", cfg.Node.ID)
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create event bus
	eventBus := events.NewEventBus()

	// Create WebSocket hub
	hub := api.NewWebSocketHub()
	go hub.Run()

	// Create agent pool
	agentPool := master.NewAgentPool(cfg)

	// Load agents from config (automations reference other nodes)
	loadAgentsFromConfig(cfg, agentPool)

	// Start health checks
	agentPool.StartHealthChecks(30 * time.Second)

	// Forward agent state changes to dashboard clients
	agentPool.OnAgentStateChange(func(nodeID string, data map[string]interface{}) {
		hub.Broadcast(data)
	})

	// Create automation engine
	autoEngine := automation.NewEngine(eventBus, manager, agentPool, database, cfg.Node.ID)
	autoEngine.Start()
	defer autoEngine.Stop()

	// Create sensor manager
	sensorManager := sensors.NewSensorManager(eventBus, g)

	// Create API handler with agent pool
	handler := api.NewHandler(cfg, manager, database, hub, agentPool, autoEngine, sensorManager)

	// Create and configure server
	port := cfg.Network.ListenPort
	if port == 0 {
		port = 8080
	}
	server := api.NewServer(port, staticFS)
	router := server.Router()

	// Add middleware
	router.Use(api.CORSMiddleware())
	router.Use(api.LoggerMiddleware())
	if cfg.Security.APIKey != "" {
		router.Use(api.APIKeyMiddleware(cfg.Security.APIKey))
	}

	// Register routes
	handler.RegisterRoutes(router)

	// Serve static files (Vue frontend)
	server.ServeStatic()

	// Handle graceful shutdown
	setupShutdown(manager)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Master server listening on %s", addr)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runAgent(cfg *config.Config, manager *node.Manager) {
	// Create agent server (thin GPIO relay — no DB, no frontend)
	agentServer := agent.NewAgentServer(cfg, manager)
	agentServer.SetupRoutes()

	// Create discovery service
	discovery := agent.NewDiscovery(cfg)

	// Start auto-discovery (register with master)
	if cfg.Network.MasterNode != "" {
		go discovery.StartRegistrationLoop()
	}

	// Connect to master via WebSocket in background
	go agentServer.ConnectToMaster()

	// Setup state change broadcasting to master
	setupAgentStateBroadcast(manager, agentServer)

	// Handle graceful shutdown
	setupShutdown(manager)

	// Start agent server
	if err := agentServer.Run(); err != nil {
		log.Fatalf("Agent server error: %v", err)
	}
}

func setupStateNotifications(manager *node.Manager, cfg *config.Config) {
	// Set up notifications for all pins
	for _, pin := range cfg.Pins {
		pinID := pin.ID
		manager.OnStateChange(pinID, func(id string, state interface{}) {
			log.Printf("Pin %s state changed to %v", id, state)
		})
	}
}

func setupAgentStateBroadcast(manager *node.Manager, agentServer *agent.AgentServer) {
	// Broadcast state changes to master
	for _, pin := range manager.GetAllStates() {
		pinID := pin.ID
		manager.OnStateChange(pinID, func(id string, state interface{}) {
			agentServer.BroadcastPinUpdate(id, state)
		})
	}
}

func loadAgentsFromConfig(cfg *config.Config, pool *master.AgentPool) {
	// Extract unique agent nodes from automation configs
	agentNodes := make(map[string]bool)
	for _, auto := range cfg.Automations {
		if auto.Trigger.Node != "" && auto.Trigger.Node != cfg.Node.ID {
			agentNodes[auto.Trigger.Node] = true
		}
		for _, action := range auto.Actions {
			if action.Node != "" && action.Node != cfg.Node.ID {
				agentNodes[action.Node] = true
			}
		}
	}

	// In real scenario, you'd load agent configs from a separate config
	log.Printf("Discovered %d agent nodes from config", len(agentNodes))
}

func setupShutdown(manager *node.Manager) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		manager.Close()
		os.Exit(0)
	}()
}


