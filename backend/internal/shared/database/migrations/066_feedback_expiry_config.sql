-- 066_feedback_expiry_config.sql
-- Add configuration for feedback link expiry and automation delay in minutes

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
(
    'feedback_link_expiry_minutes',
    '2880', -- Default 48 hours
    false,
    'Feedback Link Expiry (Minutes)',
    'marketing',
    21
),
(
    'feedback_automation_delay_minutes',
    '7200', -- Default 5 days
    false,
    'Feedback Automation Delay (Minutes)',
    'marketing',
    22
)
ON CONFLICT (key) DO NOTHING;
