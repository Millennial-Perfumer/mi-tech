-- Migration: 124_add_judgeme_posted_to_customer_feedback.sql
-- Description: Adds judgeme_posted flag and timestamp to customer_feedback table

ALTER TABLE customer_feedback 
ADD COLUMN IF NOT EXISTS judgeme_posted BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS judgeme_posted_at TIMESTAMP WITH TIME ZONE;
