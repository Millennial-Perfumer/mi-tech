-- 087_rename_config_key_to_danger_zone.sql
-- Rename show_reset_warehouse_button to enable_danger_zone for better clarity

UPDATE app_configs 
SET key = 'enable_danger_zone' 
WHERE key = 'show_reset_warehouse_button';
