-- 119_abandoned_checkouts_schema.sql
-- Schema for tracking abandoned Shopify checkouts and recovery messages

CREATE TABLE IF NOT EXISTS abandoned_checkouts (
    id SERIAL PRIMARY KEY,
    store_id TEXT NOT NULL,
    checkout_id TEXT NOT NULL,
    checkout_token TEXT NOT NULL,
    cart_token TEXT,
    email TEXT,
    phone TEXT,
    customer_name TEXT,
    checkout_url TEXT,
    line_items JSONB,
    total_price NUMERIC(15, 2),
    currency TEXT,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    order_id TEXT,
    recovery_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, PROCESSING, SENT, FAILED, CANCELLED
    recovery_attempts INTEGER DEFAULT 0,
    recovery_message_sent_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    marketing_consent BOOLEAN DEFAULT FALSE,
    sms_consent BOOLEAN DEFAULT FALSE,
    abandoned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, checkout_token)
);

CREATE INDEX idx_abandoned_checkouts_recovery ON abandoned_checkouts(completed, recovery_status, abandoned_at);
