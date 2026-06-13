-- 071_feedback_template_config.sql
-- Add configuration for the WhatsApp template used for feedback requests

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
(
    'feedback_whatsapp_template_name',
    '', -- User will set this in Settings
    false,
    'Feedback WhatsApp Template Name',
    'feedback',
    1
)
ON CONFLICT (key) DO NOTHING;

-- Move feedback_base_url to the 'feedback' category for cleaner organization
UPDATE app_configs 
SET category = 'feedback', sort_order = 2
WHERE key = 'feedback_base_url';
