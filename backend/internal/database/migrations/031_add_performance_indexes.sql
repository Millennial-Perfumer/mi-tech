-- 031_add_performance_indexes.sql
-- Add indexes to optimize dashboard and GST reporting performance.

-- 1. Index on orders(created_at) for fast date-range filtering.
-- This speeds up dashboard metrics, GST summaries, and state summaries.
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (created_at);

-- 2. Index on order_line_items(order_id) for efficient JOINs.
-- This optimizes the HSN summary query which performs heavy JOINs between orders and line items.
CREATE INDEX IF NOT EXISTS idx_order_line_items_order_id ON order_line_items (order_id);
