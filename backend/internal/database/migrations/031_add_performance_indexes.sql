-- 031_add_performance_indexes.sql
-- Add indexes to improve reporting and listing performance

-- 1. Index on orders.created_at for date-range filtered reports (GST, Dashboard, etc.)
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- 2. Index on order_line_items.order_id for faster lookups and joins
CREATE INDEX IF NOT EXISTS idx_order_line_items_order_id ON order_line_items(order_id);
