-- 072_reorganize_feedback_settings.sql
-- Move existing feedback-related settings to the dedicated 'feedback' category for better centralization

UPDATE app_configs 
SET category = 'feedback' 
WHERE key IN ('feedback_link_expiry_minutes', 'feedback_automation_delay_minutes');
