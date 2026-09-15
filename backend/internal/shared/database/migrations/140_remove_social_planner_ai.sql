-- Retire the unused Social Media, Auto Queue, Planner, and standalone AI Analysis features.
-- This is intentionally a forward migration so existing deployments remove the
-- data and foreign-key dependency without rewriting migration history.

ALTER TABLE whatsapp_conversations
    DROP COLUMN IF EXISTS active_task_id;

DROP TABLE IF EXISTS social_queue_posts;
DROP TABLE IF EXISTS social_metrics_history;
DROP TABLE IF EXISTS social_post_history;
DROP TABLE IF EXISTS social_accounts;

DROP TABLE IF EXISTS ai_memories;
DROP TABLE IF EXISTS ai_messages;
DROP TABLE IF EXISTS ai_conversations;

DROP TABLE IF EXISTS planner_task_logs;
DROP TABLE IF EXISTS planner_tasks;
DROP TABLE IF EXISTS planner_sprints;
DROP TABLE IF EXISTS planner_columns;
DROP TABLE IF EXISTS planner_boards;

DELETE FROM app_configs
WHERE key IN (
    'facebook_page_id',
    'instagram_business_id',
    'threads_user_id',
    'gdrive_automation_folder_url',
    'gdrive_service_account_json',
    'gdrive_refresh_token',
    'gdrive_client_id',
    'gdrive_client_secret',
    'gdrive_access_token',
    'n8n_webhook_url',
    'openai_api_key',
    'ai_provider',
    'ai_cloud_model',
    'ai_local_model',
    'ai_local_url',
    'ai_enabled',
    'kanban_enabled',
    'kanban_default_board_id'
);

UPDATE machine_api_keys
SET scopes = ARRAY(
    SELECT scope
    FROM unnest(scopes) AS scope
    WHERE scope NOT IN (
        'marketing:publish',
        'planner:read',
        'planner:write',
        'planner:destructive',
        'ai:read',
        'ai:write',
        'ai:destructive'
    )
)
WHERE scopes && ARRAY[
    'marketing:publish',
    'planner:read',
    'planner:write',
    'planner:destructive',
    'ai:read',
    'ai:write',
    'ai:destructive'
];
