# Pi GPIO Dashboard — Implementation Plan

## Phase 1: Foundation — Go Backend (100%)

- [x] Project scaffold (Go modules + `main.go` + `config/`)
- [x] TOML config loader (`config/loader.go`) with validation
- [x] GPIO interface + sysfs implementation (`gpio/interface.go`, `gpio/sysfs.go`)
- [x] Real GPIO via sysfs (requires root or gpio group)
- [x] Node manager (`node/manager.go`) — manages pins, handles state changes
- [x] SQLite database (`db/db.go`) with GORM — stores pin states, action logs, automation rules
- [x] Config file watcher (`config/watcher.go`) — auto-reloads on TOML changes
- [x] WebSocket Hub (`api/websocket.go`) — broadcasts events
- [x] Friendly error message for missing config file at `/etc/pi-gpio/config.toml`

## Phase 2: Single-Pi Dashboard — Vue 3 Frontend (100%)

- [x] Vue 3 + Vite + Tailwind CSS scaffold
- [x] PinCard component — Lucide icons (no emoji), live state display
- [x] Actions: toggle, pulse, blink, set
- [x] WebSocket + HTTP polling (fallback)
- [x] Log viewer (actions table)
- [x] Connection status indicator (online/offline)
- [x] Responsive grid layout
- [x] API client with `/api` prefix (`services/api.js`)

## Phase 3: Multi-Pi Network (100%)

- [x] Agent server (`agent/server.go`) — runs on each agent Pi
- [x] Agent pool (`master/agent_pool.go`) — health checks, reconnection
- [x] Agent client (`master/agent_client.go`) — WebSocket to agent
- [x] Auto-discovery (`agent/discovery.go`) — heartbeat + registration
- [x] Multi-node dashboard (`components/NodeCard.vue`)
- [x] Remote action execution via `agent_pool.ForwardAction`
- [x] Agent auto-discovery with heartbeat and health checks
- [x] Fixed: frontend API calls missing `/api` prefix

## Phase 4: Automation Engine (100%)

- [x] Rule registry (`automation/registry.go`) — JSON serialization, execution history
- [x] Trigger evaluator (`automation/trigger.go`) — `pin_state`, `value_threshold`, `timer`, `long_press`
- [x] Action execution (`automation/engine.go`) — local, forwarded, notify
- [x] Rule CRUD API (`api/routes.go`) — create, update, get, delete, enable/disable
- [x] Event bus (`events/eventbus.go`) — decoupled pub/sub
- [x] Timer-based triggers (`timerLoop`) — interval/cron support
- [x] Execution history tracking (`ExecutionRecord`)
- [x] Sysfs GPIO interface for Raspberry Pi compatibility

## Phase 5: Visual Builder (100%)

- [x] Drag-and-drop rule editor (Vue Flow)
- [x] Node types: trigger, action, condition, delay
- [x] Flow validation (no orphaned nodes, no cycles)
- [x] Live preview of rule execution
- [x] Save/load flow definitions

## Phase 6: Advanced Sensors (100%)

- [x] DHT22/AM2302 temperature/humidity (RPi GPIO + sysfs)
- [x] PIR motion sensor (digital input)
- [x] ADC (MCP3008) for analog sensors
- [x] Sensor data collection & storage
- [x] Graphing (Chart.js or similar)

## Phase 7: Polish (100%)

- [x] Tailwind CSS styling refinement
- [x] Dark mode toggle
- [x] Mobile-responsive layout
- [x] Error handling & user feedback
- [x] Documentation (README, AGENTS.md)
- [x] Production build optimization
- [x] systemd service files

## Progress Tracker

| Phase | Status | Percentage |
|-------|--------|------------|
| Phase 1: Foundation | Complete | 100% |
| Phase 2: Single-Pi Dashboard | Complete | 100% |
| Phase 3: Multi-Pi Network | Complete | 100% |
| Phase 4: Automation Engine | Complete | 100% |
| Phase 5: Visual Builder | Complete | 100% |
| Phase 6: Advanced Sensors | Complete | 100% |
| Phase 7: Polish | Complete | 100% |
| **Overall** | **Complete** | **100%** |

## Notes

- **Real GPIO**: Uses Linux sysfs interface (requires root or gpio group membership)
- **Config**: `CONFIG_PATH=/path/to/config.toml` overrides default
- **Agent**: Run with `NODE_ROLE=agent` for agent mode
- **Auto-discovery**: Agents register with master via HTTP POST + heartbeat
- **Build**: `go build` in `backend/`, `npm run build` in `frontend/`
- **Frontend**: Uses Lucide icons (never emoji)
- **API**: All endpoints under `/api` prefix
- **Automation**: Rules are JSON-serialized in SQLite, trigger evaluator supports edge detection
- **History**: Execution history tracked with duration, step-by-step results, timestamps
