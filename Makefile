.PHONY: install install-frontend install-frontend-feedback install-shop-chat-agent install-backend run frontend frontend-feedback backend build build-frontend build-frontend-feedback build-shop-chat-agent build-backend clean db-up db-down
export GOMODCACHE=$(shell pwd)/backend/.gocache/mod
export GOCACHE=$(shell pwd)/backend/.gocache/build
export GOFLAGS=-buildvcs=false
export CGO_ENABLED=0

# Install dependencies for both frontend and backend
install: install-frontend install-frontend-feedback install-shop-chat-agent install-backend

install-frontend:
	cd frontend && npm ci --legacy-peer-deps

install-frontend-feedback:
	cd frontend-feedback && npm ci --legacy-peer-deps

install-shop-chat-agent:
	cd shop-chat-agent && npm ci

install-backend:
	cd backend && go mod download

# Run both applications (backend in background, frontend in foreground)
run: db-up
	@echo "Starting backend and frontends..."
	@make backend & make frontend & make frontend-feedback & wait

# Start local PostgreSQL database container
db-up:
	@echo "Starting PostgreSQL database container..."
	cd backend && docker-compose up -d

# Stop local PostgreSQL database container
db-down:
	@echo "Stopping PostgreSQL database container..."
	cd backend && docker-compose down

frontend:
	cd frontend && npm run dev

frontend-feedback:
	cd frontend-feedback && npm run dev

backend:
	cd backend && go run github.com/air-verse/air@latest -c .air.toml

# Build both applications
build: build-frontend build-frontend-feedback build-shop-chat-agent build-backend

build-frontend:
	cd frontend && npm run build

build-frontend-feedback:
	cd frontend-feedback && npm run build

build-shop-chat-agent:
	cd shop-chat-agent && npm run build

build-backend:
	cd backend && go build -o bin/api ./cmd/main.go

# Clean build artifacts
clean:
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf frontend-feedback/dist
	rm -rf frontend-feedback/node_modules
	rm -rf backend/bin
	chmod -R +w backend/.gocache || true
	rm -rf backend/.gocache
