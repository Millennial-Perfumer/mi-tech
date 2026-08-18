-- 123_create_shopify_carts_table.sql
-- Create shopify_carts table to track cart creation and updates from Shopify webhooks

CREATE TABLE IF NOT EXISTS shopify_carts (
    id SERIAL PRIMARY KEY,
    store_id TEXT NOT NULL,
    cart_token TEXT NOT NULL UNIQUE,
    line_items JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shopify_carts_store_created ON shopify_carts(store_id, created_at);
