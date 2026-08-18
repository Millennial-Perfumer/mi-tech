-- Migration 126: Add Google Drive Automation Folder Config
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES ('gdrive_automation_folder_url', 'https://drive.google.com/drive/folders/1djXkok8cuP3efyurTd2nOwoKRo-HpEC3', false, 'Google Drive Automation Folder URL', 'auto_queue', 10)
ON CONFLICT (key) DO UPDATE SET value = 'https://drive.google.com/drive/folders/1djXkok8cuP3efyurTd2nOwoKRo-HpEC3', category = 'auto_queue';

UPDATE app_configs SET category = 'auto_queue' WHERE key = 'gdrive_automation_folder_url';
