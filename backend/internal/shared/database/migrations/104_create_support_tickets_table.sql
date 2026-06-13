-- Create support_tickets table for dedicated support ticket management
CREATE TABLE IF NOT EXISTS support_tickets (
    id SERIAL PRIMARY KEY,
    ticket_id VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) DEFAULT 'medium', -- low, medium, high, urgent
    status VARCHAR(50) DEFAULT 'open',     -- open, in-progress, resolved, closed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Migrate existing support tickets from planner_tasks to support_tickets
-- Note: We check that ticket_id is not null to identify migrated tickets
INSERT INTO support_tickets (ticket_id, title, description, priority, status, created_at, updated_at)
SELECT 
    ticket_id, 
    title, 
    description, 
    priority, 
    CASE 
        WHEN status = 'done' THEN 'resolved'
        WHEN status = 'in-progress' THEN 'in-progress'
        ELSE 'open'
    END, 
    created_at, 
    updated_at
FROM planner_tasks
WHERE ticket_id IS NOT NULL;

-- Delete migrated tasks from planner_tasks
DELETE FROM planner_tasks WHERE ticket_id IS NOT NULL;
