-- 081_migrate_reset_button_to_configs.sql
-- Move the reset button visibility toggle to the unified configuration system

-- 1. Insert into app_configs (enabling it by default as requested)
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES (
    'show_reset_button',
    'true',
    false,
    'Show Reset & Resync Button',
    'system',
    50
) ON CONFLICT (key) DO UPDATE SET 
    value = EXCLUDED.value,
    label = EXCLUDED.label,
    category = EXCLUDED.category;

-- 2. Cleanup from old table
DELETE FROM app_settings WHERE key = 'show_reset_button';
