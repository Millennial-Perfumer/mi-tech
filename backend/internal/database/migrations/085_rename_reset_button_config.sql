-- 085_rename_reset_button_config.sql
-- Rename the configuration key show_reset_button to show_reset_warehouse_button for clarity

-- If the target key already exists, delete the old one to avoid unique constraint violations
DELETE FROM app_configs 
WHERE key = 'show_reset_button' 
AND EXISTS (SELECT 1 FROM app_configs WHERE key = 'show_reset_warehouse_button');

-- If the target key does not exist, perform the rename
UPDATE app_configs 
SET key = 'show_reset_warehouse_button',
    label = 'Show Warehouse Reset Button'
WHERE key = 'show_reset_button' 
AND NOT EXISTS (SELECT 1 FROM app_configs WHERE key = 'show_reset_warehouse_button');

-- Same logic for app_settings to ensure idempotency
DELETE FROM app_settings 
WHERE key = 'show_reset_button' 
AND EXISTS (SELECT 1 FROM app_settings WHERE key = 'show_reset_warehouse_button');

UPDATE app_settings
SET key = 'show_reset_warehouse_button'
WHERE key = 'show_reset_button'
AND NOT EXISTS (SELECT 1 FROM app_settings WHERE key = 'show_reset_warehouse_button');
