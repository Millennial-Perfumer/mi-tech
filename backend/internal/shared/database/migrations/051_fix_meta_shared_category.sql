-- Migration 051: Fix Meta Shared Category
-- Corrects the category from 'marketing' to 'meta_shared' for common components.

UPDATE app_configs 
SET category = 'meta_shared' 
WHERE key IN ('meta_system_user_token', 'meta_app_id');
