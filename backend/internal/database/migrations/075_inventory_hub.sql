-- Migration: 075_inventory_hub.sql
-- Description: Creates the physical inventory master table and the external platform mapping table.

-- 1. Master Inventory Table (Warehouse Authority)
CREATE TABLE IF NOT EXISTS inventory_items (
    id SERIAL PRIMARY KEY,
    mi_sku VARCHAR(50) UNIQUE NOT NULL, -- The canonical mi-XX SKU
    title VARCHAR(255) NOT NULL,
    description TEXT, -- Imported from Shopify or entered manually
    current_stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. External Mappings Table (The Bridge)
CREATE TABLE IF NOT EXISTS inventory_mappings (
    id SERIAL PRIMARY KEY,
    inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL DEFAULT 'shopify', -- 'shopify', 'amazon', 'pos'
    external_sku VARCHAR(100) NOT NULL, -- The SKU as it appears on Shopify/Amazon
    external_variant_id VARCHAR(100), -- Critical for pushing stock levels back to Shopify
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(platform, external_sku)
);

-- Index for fast lookup during order processing
CREATE INDEX IF NOT EXISTS idx_mappings_lookup ON inventory_mappings (platform, external_sku);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_inventory_items_updated_at
    BEFORE UPDATE ON inventory_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
