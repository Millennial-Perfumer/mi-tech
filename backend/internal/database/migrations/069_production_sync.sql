-- 069_production_sync.sql
-- Synchronize structural lookup data and restore missing configuration categories

-- 1. Ensure Feedback Statuses are seeded with explicit IDs
-- This resolves the Foreign Key violation in the 'orders' table sync
DO $$
BEGIN
    INSERT INTO feedback_statuses (id, name) VALUES 
    (1, 'pending'),
    (2, 'sent'),
    (3, 'completed'),
    (4, 'expired')
    ON CONFLICT (id) DO NOTHING;

    -- Also handle conflicts on name if IDs were somehow shifted but names match
    INSERT INTO feedback_statuses (name) VALUES 
    ('pending'), ('sent'), ('completed'), ('expired')
    ON CONFLICT (name) DO NOTHING;
END $$;

-- 2. Ensure Payment (Razorpay) configuration category exists with 7 standard items
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
('razorpay_key_id', '', false, 'Razorpay Key ID', 'payment', 1),
('razorpay_key_secret', '', true, 'Razorpay Key Secret', 'payment', 2),
('razorpay_webhook_secret', '', true, 'Razorpay Webhook Secret', 'payment', 3),
('razorpay_enabled', 'false', false, 'Enable Razorpay Payments', 'payment', 4),
('razorpay_currency', 'INR', false, 'Default Currency', 'payment', 5),
('razorpay_payment_success_url', '', false, 'Payment Success Redirect URL', 'payment', 6),
('razorpay_payment_failed_url', '', false, 'Payment Failure Redirect URL', 'payment', 7)
ON CONFLICT (key) DO NOTHING;

-- 3. Ensure Kanban configuration category exists with 2 standard items
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES 
('kanban_enabled', 'true', false, 'Enable Kanban Board', 'kanban', 1),
('kanban_default_board_id', '1', false, 'Default Kanban Board ID', 'kanban', 2)
ON CONFLICT (key) DO NOTHING;
