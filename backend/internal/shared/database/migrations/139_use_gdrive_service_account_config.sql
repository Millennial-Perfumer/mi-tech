-- Migration 139: Make Google Drive service-account authentication the visible Auto Queue configuration.
-- Legacy OAuth/access-token values remain stored for rollback compatibility but are hidden from the UI.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES ('gdrive_service_account_json', '', true, 'Google Drive Service Account JSON', 'auto_queue', 11)
ON CONFLICT (key) DO UPDATE SET
    is_secret = true,
    label = 'Google Drive Service Account JSON',
    category = 'auto_queue',
    sort_order = 11;

UPDATE app_configs
SET is_secret = true
WHERE key IN ('gdrive_refresh_token', 'gdrive_client_secret', 'gdrive_access_token', 'gdrive_service_account_json');

UPDATE app_configs
SET is_secret = false,
    category = 'auto_queue',
    sort_order = 10
WHERE key = 'gdrive_automation_folder_url';
