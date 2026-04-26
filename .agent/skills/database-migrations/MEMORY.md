# MI-Tech Database Migrations Memory

## Setup
- **Location**: `backend/internal/database/migrations/`
- **Naming**: Sequential numbered SQL files: `001_create_orders.sql`, `082_add_amazon_configs.sql`
- **Execution**: Auto-run on startup via `database.InitDB()` which reads and executes all `.sql` files in order
- **ORM**: GORM AutoMigrate runs first for schema, then raw SQL migrations for data/config seeding

## Patterns
- Use `ON CONFLICT (key) DO NOTHING` for idempotent seeding (especially `app_configs`)
- Use `IF NOT EXISTS` for table/column additions
- Group related config entries by `category` (e.g., `amazon`, `shopify`, `whatsapp`)
- Mark secrets with `is_secret = true` for UI masking
- Include `sort_order` for UI display priority
- Always set `updated_at = NOW()` in seed migrations

## Key Examples
```sql
-- Config seeding pattern
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order, updated_at)
VALUES ('amazon_lwa_client_id', '', true, 'LWA Client ID', 'amazon', 10, NOW())
ON CONFLICT (key) DO NOTHING;

-- Schema addition
ALTER TABLE orders ADD COLUMN IF NOT EXISTS inventory_deducted BOOLEAN DEFAULT FALSE;
```

## Gotchas
- Never use `DROP TABLE` without explicit user approval
- Migration files are never deleted — only new ones are added
- The migration runner does NOT track which migrations have been applied (all run every startup)
- Use `ON CONFLICT DO NOTHING` to make all migrations idempotent
