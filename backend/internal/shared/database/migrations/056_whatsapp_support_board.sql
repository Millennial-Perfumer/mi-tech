-- 056_whatsapp_support_board.sql
-- Create Support Board and Issue Tracking Fields

-- 1. Create the WhatsApp Support Board
INSERT INTO planner_boards (name, description) 
VALUES ('WhatsApp Support', 'System-wide board for tracking automated issues from WhatsApp');

-- 2. Create the Columns for the Support Board (assuming board_id = 2 if the only other one was 1)
-- To be safe, we'll use a subquery for the board_id
DO $$
DECLARE
    v_board_id INTEGER;
BEGIN
    SELECT id INTO v_board_id FROM planner_boards WHERE name = 'WhatsApp Support' LIMIT 1;
    
    INSERT INTO planner_columns (board_id, name, "order") VALUES 
    (v_board_id, 'New Issue', 1),
    (v_board_id, 'In Progress', 2),
    (v_board_id, 'Waiting for Customer', 3),
    (v_board_id, 'Resolved', 4);
END $$;

-- 3. Enhance WhatsApp Chat Schema for Tracking
ALTER TABLE whatsapp_chat_messages ADD COLUMN IF NOT EXISTS is_issue BOOLEAN DEFAULT FALSE;
ALTER TABLE whatsapp_chat_messages ADD COLUMN IF NOT EXISTS priority VARCHAR(50) DEFAULT 'medium';

ALTER TABLE whatsapp_conversations ADD COLUMN IF NOT EXISTS active_task_id INTEGER REFERENCES planner_tasks(id) ON DELETE SET NULL;
ALTER TABLE whatsapp_conversations ADD COLUMN IF NOT EXISTS priority VARCHAR(50) DEFAULT 'medium';
