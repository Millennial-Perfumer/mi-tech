-- Migration 024: Add configuration to toggle 'Clear All' button on Customers tab
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES ('show_clear_customers_button', 'false', false, 'Enable Clear All Customers', 'business', 10)
ON CONFLICT (key) DO NOTHING;

-- Also add to app_settings for dual-table compatibility if needed by App.tsx fetching logic
INSERT INTO app_settings (key, value)
VALUES ('show_clear_customers_button', 'false')
ON CONFLICT (key) DO NOTHING;
