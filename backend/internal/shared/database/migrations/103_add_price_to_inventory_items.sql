-- Migration: 103_add_price_to_inventory_items.sql
-- Description: Adds a price column to inventory_items to persist synchronized variant rates.
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS price NUMERIC(10, 2) NOT NULL DEFAULT 0.00;
