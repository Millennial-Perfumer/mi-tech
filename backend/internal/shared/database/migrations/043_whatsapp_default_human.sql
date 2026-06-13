-- Set default mode to 'human' for new conversations
ALTER TABLE whatsapp_conversations ALTER COLUMN mode SET DEFAULT 'human';

-- Update existing 'auto' conversations to 'human' (Manual preference)
UPDATE whatsapp_conversations SET mode = 'human' WHERE mode = 'auto';
