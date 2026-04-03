# Setup & Development Guide

This guide covers the necessary steps to set up the project locally for development and prepare it for production.

## 🛠 Prerequisites
- Go 1.21+
- Node.js (Latest LTS)
- Docker & Docker Compose

## 🚀 Quick Start
Use the root `Makefile` for the main local workflow:

```bash
# Install dependencies
make install

# Start local PostgreSQL
make db-up

# Start database, Go API (Air), and Web Frontend
make run
```

## 🏗 Build Instructions
To build the production-ready assets:

```bash
# Build frontend and compile backend binary
make build
```

## 🧪 Running Tests
### Backend
Backend tests use Go’s `testing` package with `stretchr/testify`.
```bash
cd backend && go test ./...
```

### Frontend
```bash
cd frontend && npm run lint
cd frontend-mobile && npm run lint
```

## 📜 Repository Guidelines
- **Coding Style**: Go stays `gofmt`-formatted. React/TS uses `PascalCase` for components and screens.
- **Indentation**: 2-space for frontend, tabs for backend.
- **Commits**: Follow Conventional Commits (e.g., `feat:`, `fix:`, `chore:`).
