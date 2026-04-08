-- 062_automation_events.sql
-- Create table for managing automation events/topics

CREATE TABLE IF NOT EXISTS automation_events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    topic TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed default events
INSERT INTO automation_events (name, topic, description) VALUES
('Order Placed', 'orders/create', 'Triggered when a new order is created.'),
('Order Assigned', 'orders/assigned', 'Triggered when an order is assigned to a staff member.'),
('Order Packed / Fulfilled', 'orders/fulfilled', 'Triggered when an order is fulfilled.'),
('Order Dispatched / Picked Up', 'orders/dispatched', 'Triggered when an order is dispatched.'),
('Order Out for Delivery', 'orders/out_for_delivery', 'Triggered when an order is out for delivery.'),
('Order Delivered', 'orders/delivered', 'Triggered when an order is successfully delivered.'),
('Order Updated', 'orders/updated', 'Triggered when order details are updated.'),
('Order Cancelled', 'orders/cancelled', 'Triggered when an order is cancelled.'),
('Order Paid', 'orders/paid', 'Triggered when an order payment is captured.')
ON CONFLICT (topic) DO NOTHING;
