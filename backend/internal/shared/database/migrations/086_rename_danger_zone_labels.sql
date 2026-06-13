-- 086_rename_danger_zone_labels.sql
-- Update labels for the consolidated Danger Zone configuration

UPDATE app_configs 
SET label = 'Enable Danger Zone Actions' 
WHERE key = 'show_reset_warehouse_button';
