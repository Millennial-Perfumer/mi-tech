-- Append-only history for order and customer changes.
-- These tables intentionally do not replace the current-state tables.

CREATE TABLE IF NOT EXISTS order_events (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    external_order_id VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL DEFAULT 'system',
    actor_type VARCHAR(50) NOT NULL DEFAULT 'system',
    actor_id VARCHAR(255),
    before_data JSONB,
    after_data JSONB,
    diff_data JSONB,
    webhook_delivery_id VARCHAR(255),
    request_id VARCHAR(255),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_events_order_id ON order_events (order_id);
CREATE INDEX IF NOT EXISTS idx_order_events_external_order_id ON order_events (external_order_id);
CREATE INDEX IF NOT EXISTS idx_order_events_type ON order_events (event_type);
CREATE INDEX IF NOT EXISTS idx_order_events_occurred_at ON order_events (occurred_at DESC);

CREATE TABLE IF NOT EXISTS customer_events (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT,
    order_id BIGINT,
    customer_phone VARCHAR(50),
    event_type VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL DEFAULT 'system',
    actor_type VARCHAR(50) NOT NULL DEFAULT 'system',
    actor_id VARCHAR(255),
    before_data JSONB,
    after_data JSONB,
    diff_data JSONB,
    request_id VARCHAR(255),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customer_events_customer_id ON customer_events (customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_events_order_id ON customer_events (order_id);
CREATE INDEX IF NOT EXISTS idx_customer_events_phone ON customer_events (customer_phone);
CREATE INDEX IF NOT EXISTS idx_customer_events_type ON customer_events (event_type);
CREATE INDEX IF NOT EXISTS idx_customer_events_occurred_at ON customer_events (occurred_at DESC);
