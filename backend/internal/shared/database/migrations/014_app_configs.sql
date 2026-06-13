-- Migration 014: Create app_configs table and migrate secrets from app_settings
-- app_configs stores API keys, tokens, and configuration values.
-- The is_secret flag controls whether values are masked in API responses.

CREATE TABLE IF NOT EXISTS app_configs (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    is_secret BOOLEAN NOT NULL DEFAULT false,
    label VARCHAR(200) NOT NULL DEFAULT '',
    category VARCHAR(50) NOT NULL DEFAULT 'system',
    sort_order INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Shopify Configs
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'shopify_store_url', COALESCE(value, ''), false, 'Store URL', 'shopify', 1
FROM app_settings WHERE key = 'shopify_store_url'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'shopify_access_token', COALESCE(value, ''), true, 'Access Token', 'shopify', 2
FROM app_settings WHERE key = 'shopify_access_token'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'shopify_webhook_secret', COALESCE(value, ''), true, 'Webhook Secret', 'shopify', 3
FROM app_settings WHERE key = 'shopify_webhook_secret'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'shopify_api_version', COALESCE(value, '2026-01'), false, 'API Version', 'shopify', 4
FROM app_settings WHERE key = 'shopify_api_version'
ON CONFLICT (key) DO NOTHING;

-- WhatsApp Configs
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'whatsapp_phone_number_id', COALESCE(value, ''), false, 'Phone Number ID', 'whatsapp', 1
FROM app_settings WHERE key = 'whatsapp_phone_number_id'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'whatsapp_access_token', COALESCE(value, ''), true, 'Access Token', 'whatsapp', 2
FROM app_settings WHERE key = 'whatsapp_access_token'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'whatsapp_app_id', COALESCE(value, ''), false, 'App ID', 'whatsapp', 3
FROM app_settings WHERE key = 'whatsapp_app_id'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'whatsapp_app_secret', COALESCE(value, ''), true, 'App Secret', 'whatsapp', 4
FROM app_settings WHERE key = 'whatsapp_app_secret'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'whatsapp_waba_id', COALESCE(value, ''), false, 'WABA ID', 'whatsapp', 5
FROM app_settings WHERE key = 'whatsapp_waba_id'
ON CONFLICT (key) DO NOTHING;

-- System Configs
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'jwt_secret', COALESCE(value, ''), true, 'JWT Secret', 'system', 1
FROM app_settings WHERE key = 'jwt_secret'
ON CONFLICT (key) DO NOTHING;

-- Clean up migrated keys from app_settings
DELETE FROM app_settings WHERE key IN (
    'shopify_store_url', 'shopify_access_token', 'shopify_webhook_secret', 'shopify_api_version',
    'whatsapp_phone_number_id', 'whatsapp_access_token', 'whatsapp_app_id', 'whatsapp_app_secret', 'whatsapp_waba_id',
    'jwt_secret'
);
