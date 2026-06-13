-- Migration: Support Multiple Oils in Manufacturing
-- Created: 2026-05-01

CREATE TABLE manufacturing_oils (
    id SERIAL PRIMARY KEY,
    manufacturing_record_id INTEGER NOT NULL REFERENCES manufacturing_records(id) ON DELETE CASCADE,
    oil_inventory_id INTEGER NOT NULL REFERENCES oil_inventory(id) ON DELETE RESTRICT,
    quantity_grams DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Migrate existing data from manufacturing_records to manufacturing_oils
INSERT INTO manufacturing_oils (manufacturing_record_id, oil_inventory_id, quantity_grams, created_at, updated_at)
SELECT id, oil_inventory_id, oil_quantity_grams, created_at, updated_at 
FROM manufacturing_records 
WHERE oil_inventory_id IS NOT NULL;

-- Remove the old columns from manufacturing_records
ALTER TABLE manufacturing_records DROP COLUMN oil_inventory_id;
ALTER TABLE manufacturing_records DROP COLUMN oil_quantity_grams;
