-- Add order column to planner_tasks
ALTER TABLE planner_tasks ADD COLUMN p_order INTEGER DEFAULT 0;

-- Initialize existing tasks in order of creation per column
WITH task_ordered AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY column_id ORDER BY created_at ASC) as r_order
    FROM planner_tasks
)
UPDATE planner_tasks
SET p_order = task_ordered.r_order
FROM task_ordered
WHERE planner_tasks.id = task_ordered.id;
