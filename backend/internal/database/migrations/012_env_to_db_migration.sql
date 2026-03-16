-- Migration 012: Initialize app_settings keys for environment-to-DB migration
-- These values are seeded as empty strings for security, to be manually populated.

INSERT INTO app_settings (key, value) VALUES
('shopify_store_url', ''),
('shopify_access_token', ''),
('shopify_webhook_secret', ''),
('whatsapp_phone_number_id', ''),
('whatsapp_access_token', ''),
('whatsapp_app_id', ''),
('whatsapp_app_secret', ''),
('whatsapp_waba_id', ''),
('jwt_secret', ''),
('shopify_api_version', '2026-01')
ON CONFLICT (key) DO NOTHING;
