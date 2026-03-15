-- Index to optimize template aggregation queries in Automation Hub
-- This speeds up the GetTemplates query which joins automation_templates with aggregated automation_messages
CREATE INDEX IF NOT EXISTS idx_automation_messages_template_sent ON automation_messages (template_id, sent_at);
