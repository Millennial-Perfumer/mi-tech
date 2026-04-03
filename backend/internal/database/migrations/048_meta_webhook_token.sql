-- Migration 048: Add Meta Marketing Webhook Verify Token
-- This decouples Meta Ads webhook verification from the WhatsApp token.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
    ('meta_marketing_webhook_verify_token', '', true, 'Marketing Webhook Verify Token', 'marketing', 4)
ON CONFLICT (key) DO NOTHING;
