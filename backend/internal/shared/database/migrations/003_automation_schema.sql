-- 003_automation_schema.sql
-- Schema for WhatsApp automation and webhook status

CREATE TABLE IF NOT EXISTS automation_templates (
    id SERIAL PRIMARY KEY,
    store_id TEXT NOT NULL,
    template_name TEXT NOT NULL,
    language TEXT NOT NULL,
    category TEXT NOT NULL,
    body TEXT NOT NULL,
    header JSONB,
    footer TEXT,
    buttons JSONB,
    status TEXT DEFAULT 'pending',
    meta_template_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS automation_triggers (
    id SERIAL PRIMARY KEY,
    store_id TEXT NOT NULL,
    webhook_topic TEXT NOT NULL,
    template_id INTEGER REFERENCES automation_templates(id),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS automation_messages (
    id SERIAL PRIMARY KEY,
    store_id TEXT NOT NULL,
    template_id INTEGER REFERENCES automation_templates(id),
    order_id TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    message_id TEXT UNIQUE,
    status TEXT DEFAULT 'sent',
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS automation_whatsapp_settings (
    id SERIAL PRIMARY KEY,
    store_id VARCHAR(255) UNIQUE NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS webhook_status (
    id SERIAL PRIMARY KEY,
    topic VARCHAR(255),
    status VARCHAR(50),
    last_received TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO webhook_status (id, topic, status, last_received)
SELECT 1, 'none', 'inactive', NOW()
WHERE NOT EXISTS (SELECT 1 FROM webhook_status WHERE id = 1);
