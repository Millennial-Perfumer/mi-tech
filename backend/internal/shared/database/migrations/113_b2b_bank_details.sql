-- Migration 113: B2B Bank Details Configuration
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order) VALUES
('bank_name', 'HDFC Bank', false, 'Bank Name', 'business', 6),
('bank_account_no', '50100123456789', false, 'Account Number', 'business', 7),
('bank_ifsc', 'HDFC0001234', false, 'IFSC Code', 'business', 8),
('upi_id', 'parfumtraders@upi', false, 'UPI ID (For Payment QR)', 'business', 9)
ON CONFLICT (key) DO NOTHING;
