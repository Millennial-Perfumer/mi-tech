-- Add bulk_template_suffix to app_configs
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES ('bulk_template_suffix', '_marketing', false, 'WhatsApp Bulk Template Suffix', 'whatsapp', 100)
ON CONFLICT (key) DO NOTHING;
