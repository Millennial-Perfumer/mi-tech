-- 065_feedback_expiry.sql
-- Add timestamp to track when feedback request was sent

ALTER TABLE orders ADD COLUMN feedback_sent_at TIMESTAMP WITH TIME ZONE;
