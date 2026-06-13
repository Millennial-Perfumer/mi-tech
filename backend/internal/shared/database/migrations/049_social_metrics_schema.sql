-- Migration 049: Social Media Management (SMM) Schema

-- Social Accounts mapping (linking platforms to IDs)
CREATE TABLE IF NOT EXISTS social_accounts (
    id SERIAL PRIMARY KEY,
    platform VARCHAR(50) NOT NULL, -- 'facebook', 'instagram', 'threads'
    platform_id VARCHAR(255) NOT NULL UNIQUE, -- Page ID or User ID
    account_name VARCHAR(255),
    access_token TEXT, -- Optional platform-specific token
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Post History (archiving every social post)
CREATE TABLE IF NOT EXISTS social_post_history (
    id SERIAL PRIMARY KEY,
    platform VARCHAR(50) NOT NULL,
    post_id VARCHAR(255) NOT NULL UNIQUE, 
    content TEXT,
    media_url TEXT,
    thumbnail_url TEXT,
    permalink TEXT,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Metrics History (Daily snapshots for historical trend analysis)
CREATE TABLE IF NOT EXISTS social_metrics_history (
    id SERIAL PRIMARY KEY,
    post_id VARCHAR(255), -- NULL if account-level aggregate metric
    platform VARCHAR(50) NOT NULL,
    metric_date DATE NOT NULL,
    likes INTEGER DEFAULT 0,
    shares INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    engagement INTEGER DEFAULT 0,
    reach INTEGER DEFAULT 0,
    impressions INTEGER DEFAULT 0,
    saves INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(platform, post_id, metric_date)
);

-- Seed app_configs for SMM Discovery
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
    ('facebook_page_id', '', false, 'Facebook Page ID', 'marketing', 10),
    ('instagram_business_id', '', false, 'Instagram Business ID', 'marketing', 11),
    ('threads_user_id', '', false, 'Threads User ID', 'marketing', 12)
ON CONFLICT (key) DO NOTHING;
