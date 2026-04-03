-- Migration 044: Add Meta Marketing Configurations
-- These keys will be managed in the main Settings tab

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
    ('meta_app_id', '', false, 'Meta App ID', 'marketing', 1),
    ('meta_marketing_access_token', '', true, 'Marketing Access Token', 'marketing', 2),
    ('meta_marketing_ad_account_id', '', false, 'Ad Account ID (act_...)', 'marketing', 3)
ON CONFLICT (key) DO NOTHING;
