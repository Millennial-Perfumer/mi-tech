-- Planner Module Schema

-- Boards
CREATE TABLE IF NOT EXISTS planner_boards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Columns
CREATE TABLE IF NOT EXISTS planner_columns (
    id SERIAL PRIMARY KEY,
    board_id INTEGER NOT NULL REFERENCES planner_boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    "order" INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Sprints
CREATE TABLE IF NOT EXISTS planner_sprints (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    goal TEXT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'planned', -- planned, active, completed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tasks
CREATE TABLE IF NOT EXISTS planner_tasks (
    id SERIAL PRIMARY KEY,
    board_id INTEGER NOT NULL REFERENCES planner_boards(id) ON DELETE CASCADE,
    column_id INTEGER REFERENCES planner_columns(id) ON DELETE SET NULL,
    sprint_id INTEGER REFERENCES planner_sprints(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) DEFAULT 'medium', -- low, medium, high, urgent
    status VARCHAR(50) DEFAULT 'todo',     -- todo, in-progress, done, archived
    metadata JSONB DEFAULT '{}',           -- For order_id, customer_id, etc.
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Task Logs (Audit trail for analytics)
CREATE TABLE IF NOT EXISTS planner_task_logs (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES planner_tasks(id) ON DELETE CASCADE,
    from_column_id INTEGER REFERENCES planner_columns(id),
    to_column_id INTEGER REFERENCES planner_columns(id),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Default Board and Columns
INSERT INTO planner_boards (name, description) VALUES ('Main Board', 'Default operational board');

INSERT INTO planner_columns (board_id, name, "order") VALUES 
(1, 'To Do', 1),
(1, 'In Progress', 2),
(1, 'Review', 3),
(1, 'Done', 4);
