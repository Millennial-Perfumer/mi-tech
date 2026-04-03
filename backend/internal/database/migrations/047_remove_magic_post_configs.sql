-- Migration 047: Remove Legacy Magic Post Configurations
-- These keys are no longer needed after the feature removal.

DELETE FROM app_configs 
WHERE key IN (
    'google_ai_api_key',
    'aesthetic_storage_path',
    'aesthetic_hard_prompt'
);
