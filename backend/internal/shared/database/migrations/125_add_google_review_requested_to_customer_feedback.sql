-- Migration 125: Add Google Review Requested tracking columns to customer_feedback table

ALTER TABLE customer_feedback
ADD COLUMN IF NOT EXISTS google_review_requested BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS google_review_requested_at TIMESTAMP NULL;
