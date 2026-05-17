-- Allow 'tool' role in AI messages
ALTER TABLE ai_messages DROP CONSTRAINT IF EXISTS ai_messages_role_check;
ALTER TABLE ai_messages ADD CONSTRAINT ai_messages_role_check CHECK (role IN ('user', 'assistant', 'system', 'tool'));
