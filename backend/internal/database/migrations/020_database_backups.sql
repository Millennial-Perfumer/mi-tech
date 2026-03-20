-- Migration to create table for tracking database backups
CREATE TABLE IF NOT EXISTS database_backups (
    id SERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT, -- Stored in bytes for analytical queries
    status VARCHAR(50) NOT NULL, -- 'success', 'failed'
    remote_sync_status VARCHAR(50), -- 'success', 'failed', 'not_synced'
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster queries on recent backups
CREATE INDEX idx_database_backups_created_at ON database_backups(created_at DESC);
