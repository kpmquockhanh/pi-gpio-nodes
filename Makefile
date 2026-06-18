.PHONY: dev frontend build build-pi deploy install clean \
  up down \
  docker-up docker-down docker-build docker-logs docker-logs-backend docker-logs-frontend docker-logs-agent \
  docker-shell-backend docker-shell-frontend docker-shell-agent docker-clean docker-reset

# Development (with mock GPIO)
dev:
	@echo "Starting development servers..."
	@cd backend && MOCK_GPIO=true go run main.go &
	@cd frontend && npm run dev

# Build frontend
frontend:
	@echo "Building frontend..."
	@cd frontend && npm install && npm run build

# Build for current platform
build: frontend
	@echo "Building backend..."
	@cd backend && go build -o ../bin/pi-gpio-dashboard main.go

# Cross-compile for Raspberry Pi (ARM64)
build-pi: frontend
	@echo "Building for Raspberry Pi (ARM64)..."
	@cd backend && GOOS=linux GOARCH=arm64 go build -o ../bin/pi-gpio-dashboard-arm64 main.go

# Deploy to Pi (configure PI_HOST env var)
deploy:
	@echo "Deploying to $(PI_HOST)..."
	@scp bin/pi-gpio-dashboard-arm64 pi@$(PI_HOST):/tmp/
	@ssh pi@$(PI_HOST) "sudo mv /tmp/pi-gpio-dashboard-arm64 /usr/local/bin/pi-gpio-dashboard && sudo systemctl restart pi-gpio-dashboard"

# Install systemd service on Pi
install:
	@echo "Installing systemd service..."
	@scp backend/pi-gpio-dashboard.service pi@$(PI_HOST):/tmp/
	@ssh pi@$(PI_HOST) "sudo mv /tmp/pi-gpio-dashboard.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable pi-gpio-dashboard"

# Full release: build + deploy
release: build-pi deploy

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -rf frontend/dist/
	@rm -rf backend/static/

# Docker Development Environment
# ------------------------------

# Short aliases
up: docker-up
down: docker-down

# Start Docker dev environment (backend + frontend)
docker-up:
	@echo "Starting Docker dev environment..."
	@docker-compose up -d

# Stop Docker dev environment
docker-down:
	@echo "Stopping Docker dev environment..."
	@docker-compose down

# Rebuild Docker images
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

# View Docker logs
docker-logs:
	@docker-compose logs -f

# View backend logs only
docker-logs-backend:
	@docker-compose logs -f backend

# View frontend logs only
docker-logs-frontend:
	@docker-compose logs -f frontend

# View agent logs only
docker-logs-agent:
	@docker-compose logs -f agent

# Open shell in backend container
docker-shell-backend:
	@docker-compose exec backend sh

# Open shell in frontend container
docker-shell-frontend:
	@docker-compose exec frontend sh

# Open shell in agent container
docker-shell-agent:
	@docker-compose exec agent sh

# Clean Docker containers, images, and volumes
docker-clean:
	@echo "Cleaning Docker environment..."
	@docker-compose down -v --rmi local --remove-orphans

# Full Docker reset (includes build cache)
docker-reset: docker-clean
	@echo "Pruning Docker build cache..."
	@docker builder prune -f
