-- 064_feedback_config.sql
-- Add configuration for the customer feedback survey URL

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES (
    'feedback_base_url',
    'https://feedback-form.millennialperfumer.in',
    false,
    'Feedback Survey Base URL',
    'marketing',
    20
) ON CONFLICT (key) DO NOTHING;
