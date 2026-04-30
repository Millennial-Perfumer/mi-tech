-- Migration 089: Add Inventory Sync Configuration
-- This allows enabling/disabling automated stock updates to Shopify and Amazon.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order, updated_at)
VALUES (
    'enable_inventory_sync',
    'true', -- Default to true for production safety
    false,
    'Enable Automated Inventory Sync',
    'inventory',
    50,
    NOW()
)
ON CONFLICT (key) DO UPDATE SET 
    label = EXCLUDED.label,
    category = EXCLUDED.category,
    sort_order = EXCLUDED.sort_order;
