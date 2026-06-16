# Pi GPIO Dashboard

A multi-node GPIO control dashboard for Raspberry Pi. Control relays, LEDs, buttons, and sensors across multiple Pis from a single web interface.

![Dashboard](dashboard-screenshot.png)

## Features

- **Multi-Node Support** — Control GPIO across multiple Raspberry Pis from one dashboard
- **Real-Time Updates** — WebSocket live state updates with HTTP fallback
- **Automation Engine** — Create rules with triggers (pin state, timers, thresholds) and actions
- **Visual Builder** — Drag-and-drop automation rule editor with Vue Flow
- **Sensor Support** — DHT22 temperature/humidity, PIR motion, ADC (MCP3008)
- **Mock GPIO** — Develop without hardware using software-simulated GPIO
- **Docker Dev Environment** — Full local development with hot reload

## Quick Start

### Docker Development (Recommended)

```bash
# Clone the repository
git clone git@github.com:kpmquockhanh/pi-gpio-nodes.git
cd pi-gpio-nodes

# Start all services (backend + frontend + agent)
make docker-up

# Open dashboard
open http://localhost:5173
```

Services:
- **Dashboard**: http://localhost:5173
- **API**: http://localhost:8080
- **Agent**: http://localhost:9090

### Local Development

```bash
# Backend (with mock GPIO)
cd backend
MOCK_GPIO=true go run main.go

# Frontend (in another terminal)
cd frontend
npm install
npm run dev
```

### Production Build

```bash
# Build frontend + backend binary
make build

# Or cross-compile for Raspberry Pi (ARM64)
make build-pi
```

## Configuration

Create a TOML config file at `/etc/pi-gpio/config.toml`:

```toml
[node]
id = "master-pi"
name = "Master Controller"
role = "master"

[network]
listen_port = 8080

[security]
api_key = "your-secret-key"

[[pins]]
id = "pc-power"
name = "PC Power Relay"
bcm = 17
type = "relay"
mode = "output"

[pins.actions]
toggle = {}
pulse = { default_ms = 500, max_ms = 5000 }
```

See `AGENTS.md` for multi-node setup and systemd configuration.

## Architecture

```
Master (Port 8080)
├── Vue 3 Frontend
├── Go API Server
├── WebSocket Hub
├── Automation Engine
└── SQLite Database

Agent (Port 9090)
├── Go Agent Server
└── Local GPIO Access
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/nodes` | GET | List all nodes |
| `/api/nodes/:id/pins` | GET | Get pins for a node |
| `/api/nodes/:id/pins/:pin/action` | POST | Execute action on pin |
| `/api/automations` | GET/POST | List/Create automation rules |
| `/api/logs` | GET | View action logs |
| `/ws` | WS | WebSocket for real-time updates |

## Makefile Commands

```bash
make docker-up              # Start dev environment
make docker-down            # Stop containers
make docker-build           # Rebuild images
make docker-logs            # View logs
make docker-logs-backend    # Backend logs only
make docker-logs-agent      # Agent logs only
make docker-shell-backend   # Shell into backend
make docker-shell-agent     # Shell into agent
make docker-clean           # Clean everything
make build                  # Build for current platform
make build-pi               # Cross-compile for Pi
```

## Project Structure

```
.
├── backend/
│   ├── api/          # HTTP API + WebSocket
│   ├── agent/        # Agent server + discovery
│   ├── automation/   # Rule engine + triggers
│   ├── config/       # TOML config loader
│   ├── db/           # SQLite database
│   ├── events/       # Event bus
│   ├── gpio/         # GPIO interface (sysfs + mock)
│   ├── master/       # Agent pool + client
│   ├── node/         # Pin manager
│   └── sensors/      # DHT22 + sensor support
├── frontend/
│   ├── src/
│   │   ├── components/    # Vue components
│   │   │   ├── nodes/     # Vue Flow nodes
│   │   │   └── automation/
│   │   ├── services/      # API client
│   │   └── store/         # State management
│   └── package.json
├── docker-compose.yml
├── Makefile
└── AGENTS.md
```

## Technologies

- **Backend**: Go 1.23, Gin, GORM, SQLite, Gorilla WebSocket
- **Frontend**: Vue 3, Vite, Tailwind CSS, Pinia, Vue Flow, Lucide Icons
- **DevOps**: Docker, Docker Compose, Air (hot reload)
- **Hardware**: Raspberry Pi GPIO via sysfs

## License

MIT
