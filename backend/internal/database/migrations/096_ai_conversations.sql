-- AI Conversation History and Configuration
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL DEFAULT 'New Analysis',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata JSONB,  -- For storing tool call results, etc.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_conversations_user ON ai_conversations(user_id, updated_at DESC);
CREATE INDEX idx_ai_messages_conversation ON ai_messages(conversation_id, created_at ASC);

-- Add AI config keys
INSERT INTO app_configs (key, value, category, is_secret)
VALUES
    ('openai_api_key', '', 'ai', true),
    ('ai_provider', 'cloud', 'ai', false),
    ('ai_cloud_model', 'gpt-5.4-nano', 'ai', false),
    ('ai_local_model', 'gemma4', 'ai', false),
    ('ai_local_url', 'http://localhost:11434', 'ai', false),
    ('ai_enabled', 'true', 'ai', false)
ON CONFLICT (key) DO NOTHING;
