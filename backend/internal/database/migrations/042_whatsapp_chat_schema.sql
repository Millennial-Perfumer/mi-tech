-- WhatsApp Conversations table to track unique contacts
CREATE TABLE IF NOT EXISTS whatsapp_conversations (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    contact_name VARCHAR(255),
    last_message TEXT,
    last_message_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    mode VARCHAR(20) DEFAULT 'auto', -- 'auto' or 'human'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- WhatsApp Chat Messages table for full history
CREATE TABLE IF NOT EXISTS whatsapp_chat_messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES whatsapp_conversations(id) ON DELETE CASCADE,
    message_id VARCHAR(255) UNIQUE,
    text TEXT,
    type VARCHAR(50) DEFAULT 'text',
    direction VARCHAR(10) NOT NULL, -- 'incoming' or 'outgoing'
    sender_role VARCHAR(20), -- 'user', 'assistant', 'human'
    status VARCHAR(20) DEFAULT 'sent',
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_whatsapp_chat_messages_conversation_id ON whatsapp_chat_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_whatsapp_chat_messages_sent_at ON whatsapp_chat_messages(sent_at);

-- Add a comment explaining the purpose
COMMENT ON COLUMN whatsapp_conversations.mode IS 'Controls if the backend should automatically reply (auto) or wait for human operator (human)';
