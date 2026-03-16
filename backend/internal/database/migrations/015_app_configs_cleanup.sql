-- Migration 015: Add missing configuration keys
-- These keys were identified as missing during the unification process.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES
('send_invoice', 'true', false, 'Send WhatsApp Invoices', 'whatsapp', 10),
('pii_protection', 'false', false, 'PII Protection Mode', 'system', 10),
('whatsapp_webhook_verify_token', '', true, 'Webhook Verify Token', 'whatsapp', 11)
ON CONFLICT (key) DO NOTHING;
