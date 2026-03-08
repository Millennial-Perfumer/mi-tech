.PHONY: install install-frontend install-backend run frontend backend build build-frontend build-backend clean db-up db-down

# Install dependencies for both frontend and backend
install: install-frontend install-backend

install-frontend:
	cd frontend && npm install

install-backend:
	cd backend && go mod download

# Run both applications (backend in background, frontend in foreground)
run: db-up
	@echo "Starting backend and frontend..."
	@make backend & make frontend & wait

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

backend:
	cd backend && go run cmd/main.go

# Build both applications
build: build-frontend build-backend

build-frontend:
	cd frontend && npm run build

build-backend:
	cd backend && go build -o bin/api cmd/main.go

# Clean build artifacts
clean:
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf backend/bin
