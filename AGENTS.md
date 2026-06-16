# Pi GPIO Dashboard — Agent & Multi-Node Setup

## Running Server and Agent on the Same Machine

Yes, you can run both the **master server** and **agent** on the same Raspberry Pi. This is useful for:
- Testing multi-node setups with one device
- Running a master that also controls its own GPIO
- Having a master and one or more local agents for different pin groups

## How It Works

Each instance needs:
1. **Different config files** (different `node.id`, `node.name`, ports)
2. **Different database files** (auto-generated based on node ID)
3. **Different process IDs** (run in separate terminals or use systemd)

## Example Setup

### 1. Master Config (`/etc/pi-gpio/master.toml`)

```toml
[node]
id = "master-pi"
name = "Master Controller"
role = "master"

[network]
listen_port = 8080
master_node = ""  # Empty for master

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
set = {}

[[pins]]
id = "status-led"
name = "Status LED"
bcm = 18
type = "led"
mode = "output"

[pins.actions]
set = {}
blink = { max_times = 10, default_times = 3, interval_ms = 200 }

[[pins]]
id = "power-button"
name = "Power Button"
bcm = 27
type = "button"
mode = "input"
pull = "up"
debounce_ms = 50
```

### 2. Agent Config (`/etc/pi-gpio/agent.toml`)

```toml
[node]
id = "agent-pi"
name = "Local Agent"
role = "agent"

[network]
listen_port = 9090
master_node = "master-pi"  # Must match master node ID

[security]
api_key = "your-secret-key"  # Must match master API key

[[pins]]
id = "fan-control"
name = "Case Fan"
bcm = 22
type = "relay"
mode = "output"

[pins.actions]
toggle = {}
set = {}

[[pins]]
id = "temp-sensor"
name = "DHT22 Sensor"
bcm = 23
type = "dht22"
mode = "input"
```

## Running Both Instances

### Terminal 1 — Master
```bash
sudo CONFIG_PATH=/etc/pi-gpio/master.toml ./pi-gpio-dashboard
```

### Terminal 2 — Agent
```bash
sudo CONFIG_PATH=/etc/pi-gpio/agent.toml ./pi-gpio-dashboard
```

## What Happens

1. **Master** starts on port 8080
2. **Agent** starts on port 9090, connects to master at port 8080
3. Master sees agent in its dashboard
4. You can control agent pins from the master's web UI
5. Automation rules can trigger across both nodes

## systemd Service Files

### `/etc/systemd/system/pi-gpio-master.service`
```ini
[Unit]
Description=Pi GPIO Dashboard Master
After=network.target

[Service]
Type=simple
User=root
Environment="CONFIG_PATH=/etc/pi-gpio/master.toml"
ExecStart=/usr/local/bin/pi-gpio-dashboard
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### `/etc/systemd/system/pi-gpio-agent.service`
```ini
[Unit]
Description=Pi GPIO Dashboard Agent
After=network.target

[Service]
Type=simple
User=root
Environment="CONFIG_PATH=/etc/pi-gpio/agent.toml"
ExecStart=/usr/local/bin/pi-gpio-dashboard
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Enable and start
```bash
sudo systemctl daemon-reload
sudo systemctl enable pi-gpio-master pi-gpio-agent
sudo systemctl start pi-gpio-master pi-gpio-agent
```

## Permissions

The service needs root access or membership in the `gpio` group:

```bash
# Option 1: Run as root (easiest)
sudo ./pi-gpio-dashboard

# Option 2: Add user to gpio group
sudo usermod -a -G gpio $USER
# Log out and back in for changes to take effect
```

## Cross-Machine Setup (Tailscale)

For multiple Raspberry Pis on different networks:

1. Install Tailscale on all devices
2. Use Tailscale IPs in config
3. Firewall rules should allow the ports between Tailscale IPs

```toml
[network]
listen_port = 8080
master_node = "master-pi"
tailscale_ip = "100.x.x.x"  # Optional, for reference
```

## Troubleshooting

### "Permission denied" on GPIO
- Run as root or add user to `gpio` group
- Check `/sys/class/gpio` exists and is accessible

### Agent can't connect to master
- Verify `api_key` matches on both
- Check `master_node` ID matches exactly
- Ensure ports are not blocked by firewall
- Check `tailscale status` if using Tailscale

### Port already in use
- Each instance needs a unique `listen_port`
- Check with `sudo lsof -i :8080`

### Database conflicts
- Each node auto-creates `pi-gpio-{node_id}.db`
- Ensure node IDs are unique
