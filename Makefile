.PHONY: install install-frontend install-frontend-mobile install-frontend-feedback install-backend run frontend frontend-mobile frontend-feedback backend build build-frontend build-frontend-feedback build-backend clean db-up db-down
export GOMODCACHE=$(shell pwd)/backend/.gocache/mod
export GOCACHE=$(shell pwd)/backend/.gocache/build
export GOFLAGS=-buildvcs=false

# Install dependencies for both frontend and backend
install: install-frontend install-frontend-mobile install-frontend-feedback install-backend

install-frontend:
	cd frontend && npm install --legacy-peer-deps

install-frontend-mobile:
	cd frontend-mobile && npm install --legacy-peer-deps

install-frontend-feedback:
	cd frontend-feedback && npm install

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

frontend-mobile:
	cd frontend-mobile && npm run dev

frontend-feedback:
	cd frontend-feedback && npm run dev

backend:
	cd backend && go run github.com/air-verse/air@latest -c .air.toml

# Build both applications
build: build-frontend build-frontend-feedback build-backend

build-frontend:
	cd frontend && npm run build

build-frontend-feedback:
	cd frontend-feedback && npm run build

build-backend:
	cd backend && go build -o bin/api cmd/main.go

# Clean build artifacts
clean:
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf frontend-mobile/dist
	rm -rf frontend-mobile/node_modules
	rm -rf frontend-feedback/dist
	rm -rf frontend-feedback/node_modules
	rm -rf backend/bin
	chmod -R +w backend/.gocache || true
	rm -rf backend/.gocache
