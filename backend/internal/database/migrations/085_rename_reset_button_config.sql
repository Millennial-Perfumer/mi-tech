-- 085_rename_reset_button_config.sql
-- Rename the configuration key show_reset_button to show_reset_warehouse_button for clarity

UPDATE app_configs 
SET key = 'show_reset_warehouse_button',
    label = 'Show Warehouse Reset Button'
WHERE key = 'show_reset_button';

-- Also update app_settings just in case there are legacy entries
UPDATE app_settings
SET key = 'show_reset_warehouse_button'
WHERE key = 'show_reset_button';
