-- POS Terminals: tracks order numbering per terminal
CREATE TABLE IF NOT EXISTS pos_terminals (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,       -- 'POS1'
    name VARCHAR(100) NOT NULL,             -- 'Main Store Counter'
    next_sequence INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed the first terminal (orders start at POS1-001)
INSERT INTO pos_terminals (code, name) VALUES ('POS1', 'Main Store Counter')
ON CONFLICT (code) DO NOTHING;

-- Backfill: Create 'pos' platform mappings for all inventory items.
-- This allows syncInventoryDeltas to resolve POS line item SKUs (mi_sku)
-- through the same inventory_mappings lookup used by Shopify and Amazon.
INSERT INTO inventory_mappings (inventory_item_id, platform, external_sku)
SELECT id, 'pos', mi_sku FROM inventory_items
ON CONFLICT (platform, external_sku) DO NOTHING;
