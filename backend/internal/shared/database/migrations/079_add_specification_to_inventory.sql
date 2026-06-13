-- Migration: 079_add_specification_to_inventory
-- Description: Add specification column to inventory_items to store technical product data

ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS specification TEXT;
