-- Migration 060: Add WhatsApp Flow Private Key
-- This key is required for interactive flows data exchange.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
    ('whatsapp_flow_private_key', '', true, 'Flow Private Key (PEM format)', 'whatsapp', 10)
ON CONFLICT (key) DO NOTHING;
