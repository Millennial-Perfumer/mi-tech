-- Add admin_comment to customer_feedback table
ALTER TABLE customer_feedback
ADD COLUMN IF NOT EXISTS admin_comment TEXT;
