-- Migration: 080_inventory_logs.sql
-- Description: Creates the inventory_logs table for audit trails and stock tracking.

CREATE TABLE IF NOT EXISTS inventory_logs (
    id SERIAL PRIMARY KEY,
    inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE CASCADE,
    delta INTEGER NOT NULL, -- positive for additions, negative for deductions
    reason VARCHAR(50) NOT NULL, -- 'sale', 'cancellation', 'return', 'lost', 'manual', 'correction'
    platform VARCHAR(50) NOT NULL, -- 'internal', 'shopify', 'amazon'
    external_order_id VARCHAR(100), -- Optional: Link to a specific order
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast audit review
CREATE INDEX IF NOT EXISTS idx_inventory_logs_item ON inventory_logs (inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_logs_order ON inventory_logs (external_order_id);
