# MI-Tech Docker & Deployment Memory

## Local Development
- **Database**: PostgreSQL 16 via `backend/docker-compose.yml` (`make db-up`)
- **Backend**: Go with Air hot-reload (`make backend`)
- **Frontend**: Vite dev server (`make frontend`)
- **All together**: `make run` (starts DB, backend, frontend)

## Production
- **Docker Compose**: `docker-compose.prod.yml` at project root
- **Backend Dockerfile**: `backend/Dockerfile` — multi-stage build
- **Nginx**: Reverse proxy config in `nginx/`
- **Monitoring**: Prometheus + Grafana setup in `monitoring/` with `docker-compose.monitoring.yml`
- **SSL**: `init-ssl.sh` for Let's Encrypt setup

## Build
- Backend: `cd backend && go build -o bin/api cmd/main.go`
- Frontend: `cd frontend && npm run build` → outputs to `frontend/dist/`
- Frontend Mobile: `cd frontend-mobile && npm run build`
- Frontend Feedback: `cd frontend-feedback && npm run build`

## Key Environment Variables
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
- `PORT` (default 8080)
- Amazon/Shopify/WhatsApp keys (now in `app_configs` DB table with ENV fallback)
- `JWT_SECRET`
