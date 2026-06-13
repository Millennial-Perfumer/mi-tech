-- Migration 082: Add Amazon SP-API Configurations
INSERT INTO app_configs (key, value, is_secret, label, category, sort_order, updated_at) VALUES
('amazon_lwa_client_id', '', true, 'LWA Client ID', 'amazon', 10, NOW()),
('amazon_lwa_client_secret', '', true, 'LWA Client Secret', 'amazon', 20, NOW()),
('amazon_lwa_refresh_token', '', true, 'LWA Refresh Token', 'amazon', 30, NOW()),
('amazon_aws_access_key', '', true, 'AWS Access Key', 'amazon', 40, NOW()),
('amazon_aws_secret_key', '', true, 'AWS Secret Key', 'amazon', 50, NOW()),
('amazon_aws_region', 'eu-west-1', false, 'AWS Region', 'amazon', 60, NOW()),
('amazon_aws_role_arn', '', false, 'AWS Role ARN', 'amazon', 70, NOW()),
('amazon_marketplace_id', 'A21TJRUUN4KGV', false, 'Marketplace ID', 'amazon', 80, NOW()),
('amazon_seller_id', '', false, 'Seller ID', 'amazon', 90, NOW())
ON CONFLICT (key) DO NOTHING;
