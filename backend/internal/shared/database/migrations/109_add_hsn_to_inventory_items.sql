-- Migration 109: Add HSN Code to Inventory Items
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS hsn_code VARCHAR(8) DEFAULT '33029019';
