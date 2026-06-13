-- 059_feedback_system.sql
-- Look-up table for feedback states and integration into orders

CREATE TABLE IF NOT EXISTS feedback_statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO feedback_statuses (name) VALUES 
('pending'),
('sent'),
('completed'),
('expired')
ON CONFLICT (name) DO NOTHING;

-- Update orders table to track delivery and feedback status
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS feedback_status_id INTEGER REFERENCES feedback_statuses(id) DEFAULT 1;

-- Table to store actual customer feedback/ratings
CREATE TABLE IF NOT EXISTS customer_feedback (
    id SERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    customer_phone VARCHAR(50),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for the automation worker to efficiently find orders ready for feedback
CREATE INDEX IF NOT EXISTS idx_orders_delivery_feedback ON orders (delivery_status, feedback_status_id, delivered_at);
