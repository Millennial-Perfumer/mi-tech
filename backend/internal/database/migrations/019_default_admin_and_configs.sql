-- Migration 019: Seed default admin and ensure configuration keys exist
-- This ensures that fresh installations have an admin user and visible settings in the UI.

-- 1. Insert default admin user if it doesn't exist
-- Username: admin, Password: (provided by user)
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2b$12$wfxnb3A28AzabsodKXnEfO1cbs.93JIppsbXdexZkapTrEo79Mr5W')
ON CONFLICT (username) DO NOTHING;

-- 2. Ensure all essential app_configs keys exist so they show up in the UI
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
('jwt_secret', '', true, 'JWT Secret', 'system', 1),
('shopify_store_url', '', false, 'Store URL', 'shopify', 1),
('shopify_access_token', '', true, 'Access Token', 'shopify', 2),
('shopify_webhook_secret', '', true, 'Webhook Secret', 'shopify', 3),
('shopify_api_version', '2024-10', false, 'API Version', 'shopify', 4),
('whatsapp_phone_number_id', '', false, 'Phone Number ID', 'whatsapp', 1),
('whatsapp_access_token', '', true, 'Access Token', 'whatsapp', 2),
('whatsapp_app_id', '', false, 'App ID', 'whatsapp', 3),
('whatsapp_app_secret', '', true, 'App Secret', 'whatsapp', 4),
('whatsapp_waba_id', '', false, 'WABA ID', 'whatsapp', 5),
('pii_protection', 'false', false, 'PII Protection Mode', 'system', 10),
('send_invoice', 'true', false, 'Send WhatsApp Invoices', 'whatsapp', 10),
('whatsapp_webhook_verify_token', '', true, 'Webhook Verify Token', 'whatsapp', 11),
('business_name', '', false, 'Business Name', 'business', 1),
('business_gstin', '', false, 'GSTIN', 'business', 2),
('business_address_line1', '', false, 'Address Line 1', 'business', 3),
('business_address_line2', '', false, 'Address Line 2', 'business', 4),
('business_phone', '', false, 'Phone Number', 'business', 5),
('show_sync_button', 'true', false, 'Enable Manual Sync', 'system', 2),
('show_reset_button', 'false', false, 'Enable Reset Button', 'system', 3)
ON CONFLICT (key) DO NOTHING;
