-- 014_meta_first_templates.sql
-- Add variable_mappings column to automation_templates

ALTER TABLE automation_templates
ADD COLUMN IF NOT EXISTS variable_mappings JSONB DEFAULT '{}'::jsonb;
