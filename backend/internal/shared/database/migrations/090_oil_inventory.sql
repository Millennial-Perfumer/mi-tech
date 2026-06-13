-- Migration: Add Oil Inventory and Suppliers
-- Created: 2026-05-01

CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_info TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE oil_inventory (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE SET NULL,
    purchase_price_per_kg DECIMAL(10, 2) DEFAULT 0,
    grams_left DECIMAL(10, 2) DEFAULT 0,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX idx_oil_inventory_item_id ON oil_inventory(inventory_item_id);
CREATE INDEX idx_oil_inventory_supplier_id ON oil_inventory(supplier_id);
