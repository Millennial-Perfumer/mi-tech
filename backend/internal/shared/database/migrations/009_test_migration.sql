-- 009_test_migration.sql
-- Test migration to verify the tracking system
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
