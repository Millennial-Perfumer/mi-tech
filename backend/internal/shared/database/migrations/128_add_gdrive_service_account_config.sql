-- Migration 128: Add Service Account JSON Config
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES ('gdrive_service_account_json', '', true, 'Google Cloud Service Account JSON Key', 'auto_queue', 11)
ON CONFLICT (key) DO UPDATE SET category = 'auto_queue';
