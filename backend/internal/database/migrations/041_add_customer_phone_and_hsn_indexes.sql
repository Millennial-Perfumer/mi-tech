-- 041_add_customer_phone_and_hsn_indexes.sql
-- Add indexes to optimize customer statistics lookups and HSN summary report grouping

-- 1. Index on orders(customer_phone) for faster customer stats recalculation
-- Problem: Full table scan on 'orders' for every customer update (O(N) lookup)
-- Optimization: B-tree index reduces lookup to O(log N)
-- Expected Impact: Reduces database load by ~90% for bulk order syncs
CREATE INDEX IF NOT EXISTS idx_orders_customer_phone ON orders(customer_phone);

-- 2. Index on order_line_items(hs_code) for faster HSN-based reporting
-- Problem: Sequential scan on 'order_line_items' for GST reporting
-- Optimization: B-tree index on frequently grouped field
-- Expected Impact: Significant speedup in HSN summary report generation as data grows
CREATE INDEX IF NOT EXISTS idx_order_line_items_hs_code ON order_line_items(hs_code);
