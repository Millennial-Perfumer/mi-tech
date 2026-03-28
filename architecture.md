# GST Invoice Manager - System Architecture

This document serves as the primary technical reference and "brain" for the GST Invoice Manager repository. It provides an end-to-end overview of the system, components, data flows, and integrations.

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

## 4. Database Schema
A migration-based PostgreSQL database (`backend/internal/database/migrations/`).
Core tables include:
- `sources`: e.g., 'shopify', 'amazon', 'pos'.
- `orders`: Stores transaction data (total price, statuses, customer details).
- `order_line_items`: Individual products in an order, mapping `hs_code` for detailed HSN and GST splitting.
- `customers`: Aggregated buyer profiles synced from Shopify.
- `users`: System administrators and read-only users.
- `app_settings` / `app_configs`: Stores encrypted keys (Shopify, Meta) and app-wide preferences.
- `webhook_events`: Logs incoming webhooks for replay and processing.
- `whatsapp_templates`: Approved Meta message templates.
- `whatsapp_triggers`: Rules mapped to order statuses (e.g. 'Paid') that trigger a message.
- `whatsapp_messages_index`: Log of sent messages for tracking delivery status.

## 5. Core Workflows

### A. Order Ingestion
1. **Real-time (Webhooks)**: Shopify posts JSON payloads to `/api/webhooks/shopify` on events like `order/create`. Payload is verified via HMAC and saved to `webhook_events`. The payload is mapped to `orders` and `order_line_items`.
2. **Bulk Sync**: Admin requests a sync via `/api/shopify/sync`. The Go backend fetches historical data from Shopify's REST/GraphQL APIs and persists it.

### B. WhatsApp Automation
1. The backend listens for order status changes via webhooks.
2. The `messages_service` evaluates defined `whatsapp_triggers` against the incoming event.
3. If a match occurs, the service maps order data into the designated `whatsapp_template`.
4. The message is dispatched to the Meta Cloud API.
5. Webhooks from Meta arrive at `/api/automation/whatsapp/webhook` to update message delivery statuses (Sent, Delivered, Read).

### C. GST Reporting
1. The frontend requests reports via `/api/reports/summary`, `/api/reports/state-wise`, or `/api/reports/hsn-wise`.
2. The `report_repository` executes complex SQL aggregations grouped by state or HSN codes to calculate IGST, CGST, and SGST accurately.

## 6. API Routing Summary
Defined in `backend/internal/server/router.go`:
- **Auth**: `/api/auth/login` (JWT based)
- **Users/Customers**: CRUD endpoints for access and buyer info.
- **Orders**: `/api/orders`, `/api/orders/status`, `/api/orders/invoice`
- **Sync**: `/api/shopify/sync`, `/api/shopify/reset`
- **Reports**: `/api/reports/summary`, `.../state-wise`, `.../hsn-wise`
- **Automation**: `/api/automation/whatsapp/...` (Templates, Triggers, Messages, Webhook)
- **Settings/Configs**: `/api/settings`, `/api/configs`