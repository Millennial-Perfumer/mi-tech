-- 041_add_customer_phone_and_hsn_indexes.sql
-- Add indexes to improve performance of customer lookups and HSN reports

-- 1. Index on orders(customer_phone) for faster customer stats calculation (O(log N) lookup)
CREATE INDEX IF NOT EXISTS idx_orders_customer_phone ON orders(customer_phone);

-- 2. Index on order_line_items(hs_code) for faster HSN summary reports (faster grouping/aggregation)
CREATE INDEX IF NOT EXISTS idx_order_line_items_hs_code ON order_line_items(hs_code);
