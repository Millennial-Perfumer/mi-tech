-- Enrich the existing inventory movement ledger with order context and
-- enough state to explain the stock transition.

ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS order_id BIGINT;
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS customer_id BIGINT;
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS stock_before INTEGER;
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS stock_after INTEGER;
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50);
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS actor_id VARCHAR(255);
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS request_id VARCHAR(255);
ALTER TABLE inventory_logs ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_inventory_logs_order_id ON inventory_logs (order_id);
CREATE INDEX IF NOT EXISTS idx_inventory_logs_customer_id ON inventory_logs (customer_id);
