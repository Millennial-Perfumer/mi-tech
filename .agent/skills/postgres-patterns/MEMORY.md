# MI-Tech PostgreSQL Memory

## Database
- **Engine**: PostgreSQL 16 (Docker container via `backend/docker-compose.yml`)
- **ORM**: GORM with AutoMigrate + raw SQL migrations
- **Migrations**: `backend/internal/database/migrations/` — sequential numbered SQL files (e.g., `082_add_amazon_configs.sql`)
- **Connection**: DSN built from `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT` env vars

## Key Tables
- `orders` — Shopify + Amazon orders with `source_id` discriminator, `inventory_deducted` flag
- `line_items` — Order line items with SKU, quantity, price
- `inventory_items` — Canonical warehouse inventory with `mi_sku`, `shopify_sku`, `amazon_sku`
- `inventory_logs` — Audit trail for stock adjustments (reason, quantity delta, reference)
- `app_configs` — Key-value configuration store with `category`, `is_secret`, `label`, `sort_order`
- `app_settings` — Key-value settings store (date ranges, preferences)
- `customers` — Customer records linked to orders
- `webhook_events` — Shopify webhook delivery log
- `automation_messages` — WhatsApp message tracking
- `planner_items` — Kanban board items

## Patterns
- `ON CONFLICT (key) DO NOTHING` for idempotent config seeding
- `app_configs` categories: `business`, `shopify`, `amazon`, `meta_shared`, `marketing`, `social_media`, `whatsapp`, `feedback`, `system`
- Soft deletes via GORM's `DeletedAt` on some models
- Composite unique constraints on `(order_id, sku)` for line items
