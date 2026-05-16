-- 098_ai_memories.sql
-- Persistent memory for AI business rules and decisions

CREATE TABLE IF NOT EXISTS ai_memories (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    category VARCHAR(100) DEFAULT 'general', -- business_rule, analysis_logic, user_preference
    metadata JSONB, -- store additional context if needed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ai_memories_key ON ai_memories(key);
CREATE INDEX idx_ai_memories_category ON ai_memories(category);
