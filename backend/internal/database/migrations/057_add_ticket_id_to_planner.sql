-- Add ticket_id column to planner_tasks for tracking support tickets
ALTER TABLE planner_tasks ADD COLUMN ticket_id VARCHAR(50);

-- Ensure ticket_id is unique across tasks, but allow NULL for non-ticket tasks
CREATE UNIQUE INDEX idx_planner_tasks_ticket_id ON planner_tasks(ticket_id) WHERE ticket_id IS NOT NULL;
