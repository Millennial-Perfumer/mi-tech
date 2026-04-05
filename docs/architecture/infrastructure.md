# Infrastructure & Ops Reference

This document tracks the persistent state of infrastructure, deployment configurations, and reliability metrics.

## 🕒 Current Infrastructure State
- **Primary Orchestrator**: Docker Compose (version 3.8).
- **Production Manifest**: [`docker-compose.prod.yml`](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/docker-compose.prod.yml)
- **Monitoring Manifest**: [`docker-compose.monitoring.yml`](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/docker-compose.monitoring.yml)
- **Gateway**: Nginx (Alpine) with sidecar Certbot for auto-renewal every 12h.
- **Registry**: GitHub Container Registry (ghcr.io) via `deploy.yml`.

## ✅ Stability & Resilience
- **Restart Policies**: All production services use `restart: always`.
- **Database**: PostgreSQL 15-alpine with persistent volume `postgres_data_prod`.

## 🛠️ Known Quirks & Manual Steps
- **SSL Initialisation**: Initial SSL setup for new environments requires a manual run of `init-ssl.sh`.
- **Startup Sync**: The backend currently relies on `depends_on` without application-level health checks.

## 🛡️ Rollback Framework
- **Manual**: Rollback to previous Docker image tag via GHCR reference.
- **Auto**: Failed build CI smoke tests prevent image tag update.
