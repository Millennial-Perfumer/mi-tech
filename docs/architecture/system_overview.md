# System Architecture Overview

This document provides a technical overview of the GST Invoice Manager, covering the stack, directory structure, and core components.

## 1. System Overview
The GST Invoice Manager is a full-stack containerized web application designed to bridge e-commerce platforms (primarily Shopify) with automated customer communication (WhatsApp) while generating compliant GST reports. It handles order ingestion, invoice generation, real-time status-triggered messaging, and tax reporting.

## 2. Tech Stack
- **Frontend**: React 18, TypeScript, Vite
- **Backend**: Go (Golang) 1.21+
- **Database**: PostgreSQL 15 (Alpine)
- **Infrastructure**: Docker & Docker Compose
- **Reverse Proxy**: Nginx (with Let's Encrypt / Certbot for SSL)
- **Monitoring & Observability**: Grafana, Prometheus, Promtail, Loki, Tempo

## 3. Directory Structure
- `/frontend/` - React SPA (Vite + TypeScript). Contains UI for Dashboards, Configurations, WhatsApp Automation rules, and GST Reports.
- `/backend/` - Go REST API. Follows layered architecture:
  - `cmd/`: Entry points for the application and seeding scripts.
  - `internal/handler/`: HTTP transport layer (parsing requests, auth validation).
  - `internal/service/`: Core business logic (Orders, Sync, Reports, WhatsApp).
  - `internal/repository/`: Database interactions (PostgreSQL/GORM).
  - `internal/entity/`: Data structures representing DB tables.
- `/nginx/` - Nginx configuration and routing rules.
- `/monitoring/` - Configuration for Grafana, Loki (logs), Prometheus (metrics), Promtail, and Tempo (tracing).

## 4. Database Schema (High-Level)
For detailed schema, see [Database Schema Reference](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/docs/database/schema.md).
Core tables include:
- `sources`: e.g., 'shopify', 'amazon', 'pos'.
- `orders`: Stores transaction data.
- `order_line_items`: Individual products with HSN and GST splitting.
- `customers`: Aggregated buyer profiles.
- `app_settings`: Encrypted API keys (Shopify, Meta).
- `whatsapp_triggers`: Status-based automation rules.

## 5. API Routing
Defined in `backend/internal/server/router.go`:
- **Auth**: `/api/auth/login` (JWT based)
- **Orders**: `/api/orders`, `/api/orders/invoice`
- **Sync**: `/api/shopify/sync`
- **Automation**: `/api/automation/whatsapp/...`
- **Reports**: `/api/reports/summary`, `.../state-wise`, `.../hsn-wise`
