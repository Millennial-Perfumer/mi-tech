-- Migration 050: Refactor Meta Configurations
-- This unifies the Meta System User Token and reorganizes Social Media vs Marketing settings.

-- 1. Create the new unified Meta System User Token if it doesn't exist
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES ('meta_system_user_token', '', true, 'Meta System User Token (Shared)', 'meta_shared', 2)
ON CONFLICT (key) DO NOTHING;

-- 2. Migrate existing value from meta_marketing_access_token (priority) or whatsapp_access_token
UPDATE app_configs 
SET value = COALESCE(
    (SELECT value FROM app_configs WHERE key = 'meta_marketing_access_token' AND value <> ''),
    (SELECT value FROM app_configs WHERE key = 'whatsapp_access_token' AND value <> ''),
    ''
)
WHERE key = 'meta_system_user_token';

-- 3. Update existing configs to new categories and labels
UPDATE app_configs SET category = 'meta_shared', label = 'Meta App ID (Shared)', sort_order = 1 WHERE key = 'meta_app_id';
UPDATE app_configs SET category = 'marketing', label = 'Ad Account ID (act_...)', sort_order = 1 WHERE key = 'meta_marketing_ad_account_id';
UPDATE app_configs SET category = 'marketing', label = 'Marketing Webhook Verify Token', sort_order = 2 WHERE key = 'meta_marketing_webhook_verify_token';

-- Social Media category
UPDATE app_configs SET category = 'social_media', label = 'Facebook Page ID', sort_order = 1 WHERE key = 'facebook_page_id';
UPDATE app_configs SET category = 'social_media', label = 'Instagram Business ID', sort_order = 2 WHERE key = 'instagram_business_id';
UPDATE app_configs SET category = 'social_media', label = 'Threads User ID', sort_order = 3 WHERE key = 'threads_user_id';

-- WhatsApp category cleanup
UPDATE app_configs SET sort_order = 1 WHERE key = 'whatsapp_phone_number_id';
UPDATE app_configs SET sort_order = 2 WHERE key = 'whatsapp_waba_id';
UPDATE app_configs SET sort_order = 3 WHERE key = 'whatsapp_webhook_verify_token';

-- 4. Delete old tokens
DELETE FROM app_configs WHERE key IN ('meta_marketing_access_token', 'whatsapp_access_token');
