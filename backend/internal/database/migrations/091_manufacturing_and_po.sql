-- Create purchase_orders table
CREATE TABLE IF NOT EXISTS purchase_orders (
    id SERIAL PRIMARY KEY,
    oil_inventory_id INTEGER NOT NULL REFERENCES oil_inventory(id) ON DELETE CASCADE,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    quantity_grams DOUBLE PRECISION NOT NULL,
    unit_price_per_kg DOUBLE PRECISION NOT NULL,
    total_price DOUBLE PRECISION NOT NULL,
    purchase_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create manufacturing_records table
CREATE TABLE IF NOT EXISTS manufacturing_records (
    id SERIAL PRIMARY KEY,
    oil_inventory_id INTEGER NOT NULL REFERENCES oil_inventory(id) ON DELETE CASCADE,
    oil_quantity_grams DOUBLE PRECISION NOT NULL,
    manufacturing_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create manufacturing_products table
CREATE TABLE IF NOT EXISTS manufacturing_products (
    id SERIAL PRIMARY KEY,
    manufacturing_record_id INTEGER NOT NULL REFERENCES manufacturing_records(id) ON DELETE CASCADE,
    inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    quantity_produced INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_po_oil_inventory_id ON purchase_orders(oil_inventory_id);
CREATE INDEX IF NOT EXISTS idx_po_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX IF NOT EXISTS idx_mfg_oil_inventory_id ON manufacturing_records(oil_inventory_id);
CREATE INDEX IF NOT EXISTS idx_mfg_prod_mfg_record_id ON manufacturing_products(manufacturing_record_id);
CREATE INDEX IF NOT EXISTS idx_mfg_prod_inventory_item_id ON manufacturing_products(inventory_item_id);
