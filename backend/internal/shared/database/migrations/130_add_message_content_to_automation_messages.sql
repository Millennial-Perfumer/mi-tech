-- 130_add_message_content_to_automation_messages.sql
-- Store the rendered message text and the raw Meta Cloud API payload
-- so message logs can show exactly what was sent to the customer.

ALTER TABLE automation_messages ADD COLUMN IF NOT EXISTS message_text TEXT;
ALTER TABLE automation_messages ADD COLUMN IF NOT EXISTS payload JSONB;