-- 100_feedback_auto_trigger_config.sql
-- Add configuration for automatic daily feedback triggering

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
(
    'feedback_auto_trigger_enabled',
    'false',
    false,
    'Enable Auto Feedback Sending',
    'feedback',
    3
),
(
    'feedback_auto_trigger_time',
    '10:00',
    false,
    'Auto Feedback Sending Time (HH:MM)',
    'feedback',
    4
)
ON CONFLICT (key) DO NOTHING;
