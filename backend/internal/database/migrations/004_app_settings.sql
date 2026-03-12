-- 004_app_settings.sql
-- Key-value settings table for persisting user preferences

CREATE TABLE IF NOT EXISTS app_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Default date range: YTD
INSERT INTO app_settings (key, value) VALUES ('date_range_start', '') ON CONFLICT (key) DO NOTHING;
INSERT INTO app_settings (key, value) VALUES ('date_range_end', '') ON CONFLICT (key) DO NOTHING;
