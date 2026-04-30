-- 084_warehouse_reset_config.sql
-- Ensure the warehouse reset toggle is correctly configured in the app_configs table

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order, updated_at)
VALUES (
    'show_reset_warehouse_button',
    'false', -- Default to false for safety
    false,
    'Show Warehouse Reset Button',
    'system',
    100,
    NOW()
)
ON CONFLICT (key) DO UPDATE SET 
    label = EXCLUDED.label,
    category = EXCLUDED.category,
    sort_order = EXCLUDED.sort_order;
