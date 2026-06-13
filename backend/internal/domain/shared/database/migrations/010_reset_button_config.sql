-- 010_reset_button_config.sql
-- Setting to toggle the visibility of the Reset & Sync button
INSERT INTO app_settings (key, value) VALUES ('show_reset_button', 'false') ON CONFLICT (key) DO NOTHING;
