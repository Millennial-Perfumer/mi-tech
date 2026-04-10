-- Rename the support board to match the new UI module
UPDATE planner_boards SET name = 'Support Tickets' WHERE name = 'WhatsApp Support';

-- Rename columns for better ticketing feel if needed
-- 'New Issue' -> 'Open', 'In Progress' -> 'Working On', etc.
-- Actually, the user liked 'issue, resolved, working on'
DO $$
DECLARE
    v_board_id INTEGER;
BEGIN
    SELECT id INTO v_board_id FROM planner_boards WHERE name = 'Support Tickets' LIMIT 1;
    
    UPDATE planner_columns SET name = 'New Issue' WHERE board_id = v_board_id AND name = 'New Issue';
    UPDATE planner_columns SET name = 'Working On' WHERE board_id = v_board_id AND name = 'In Progress';
    UPDATE planner_columns SET name = 'Resolved' WHERE board_id = v_board_id AND name = 'Resolved';
END $$;
