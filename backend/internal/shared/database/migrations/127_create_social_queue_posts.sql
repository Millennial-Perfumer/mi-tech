-- Migration 127: Create social_queue_posts table
CREATE TABLE IF NOT EXISTS social_queue_posts (
    id SERIAL PRIMARY KEY,
    folder_name VARCHAR(255) NOT NULL UNIQUE,
    gdrive_folder_id VARCHAR(255) DEFAULT '',
    g_drive_folder_id VARCHAR(255) DEFAULT '',
    post_type VARCHAR(50) NOT NULL DEFAULT 'SINGLE_PHOTO',
    caption TEXT DEFAULT '',
    hashtags TEXT DEFAULT '',
    media_filenames TEXT DEFAULT '[]',
    target_platforms TEXT DEFAULT '[]',
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED',
    error_message TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE social_queue_posts ADD COLUMN IF NOT EXISTS gdrive_folder_id VARCHAR(255) DEFAULT '';
ALTER TABLE social_queue_posts ADD COLUMN IF NOT EXISTS g_drive_folder_id VARCHAR(255) DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_social_queue_posts_status ON social_queue_posts(status);
