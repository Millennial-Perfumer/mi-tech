-- Migration 016: Final Settings Unification
-- Move business profile and feature toggles to app_configs

-- Business Profile
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'business_name', value, false, 'Business Name', 'business', 1 FROM app_settings WHERE key = 'business_name'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'business_gstin', value, false, 'GSTIN', 'business', 2 FROM app_settings WHERE key = 'business_gstin'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'business_address_line1', value, false, 'Address Line 1', 'business', 3 FROM app_settings WHERE key = 'business_address_line1'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'business_address_line2', value, false, 'Address Line 2', 'business', 4 FROM app_settings WHERE key = 'business_address_line2'
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT 'business_phone', value, false, 'Phone Number', 'business', 5 FROM app_settings WHERE key = 'business_phone'
ON CONFLICT (key) DO NOTHING;

-- Cleanup migrated keys
DELETE FROM app_settings WHERE key IN (
    'business_name', 'business_gstin', 'business_address_line1', 'business_address_line2', 'business_phone'
);
