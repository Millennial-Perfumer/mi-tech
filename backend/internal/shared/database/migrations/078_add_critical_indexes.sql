-- 078_add_critical_indexes.sql
-- Add critical indexes for query optimization
-- Addresses N+1 queries and slow lookups

-- 1. Index on customer_phone - used for order lookups and customer stats
CREATE INDEX IF NOT EXISTS idx_orders_customer_phone ON orders(customer_phone);

-- 2. Index on external_order_id with source_id - composite for faster lookups
CREATE INDEX IF NOT EXISTS idx_orders_source_external ON orders(source_id, external_order_id);

-- 3. Index on delivery_status - used for feedback queries
CREATE INDEX IF NOT EXISTS idx_orders_delivery_status ON orders(delivery_status);

-- 4. Index on financial_status - used for filtering
CREATE INDEX IF NOT EXISTS idx_orders_financial_status ON orders(financial_status);

-- 5. Index on fulfillment_status - used for filtering
CREATE INDEX IF NOT EXISTS idx_orders_fulfillment_status ON orders(fulfillment_status);

-- 6. Index on customer_email - used for search
CREATE INDEX IF NOT EXISTS idx_orders_customer_email ON orders(customer_email);

-- 7. Index on order_number - used for search
CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);

-- 8. Index on tracking_number - used for lookup
CREATE INDEX IF NOT EXISTS idx_orders_tracking_number ON orders(tracking_number);

-- 9. Index on feedback_status_id - used for feedback queries
CREATE INDEX IF NOT EXISTS idx_orders_feedback_status ON orders(feedback_status_id);

-- 10. Index on delivered_at - used for feedback delay queries
CREATE INDEX IF NOT EXISTS idx_orders_delivered_at ON orders(delivered_at);

-- 11. Index on order_line_items.sku - used for inventory lookups
CREATE INDEX IF NOT EXISTS idx_order_line_items_sku ON order_line_items(sku);

-- 12. Composite index for common query patterns (date range + status)
CREATE INDEX IF NOT EXISTS idx_orders_created_status ON orders(created_at, status);

-- 13. Index on customers.external_id - used for customer link
CREATE INDEX IF NOT EXISTS idx_customers_external_id ON customers(external_id);

-- 14. Index on customers.phone_number - used for customer lookup
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone_number);
