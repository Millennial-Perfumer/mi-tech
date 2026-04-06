-- 054_delhivery_schema.sql
-- Schema for Delhivery courier integration

CREATE TABLE IF NOT EXISTS delhivery_configs (
    id SERIAL PRIMARY KEY,
    api_token TEXT NOT NULL,
    warehouse_name TEXT NOT NULL DEFAULT 'Parfum Traders',
    env TEXT NOT NULL DEFAULT 'staging', -- 'staging' or 'production'
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS package_matrix (
    id SERIAL PRIMARY KEY,
    product_count INTEGER UNIQUE NOT NULL,
    height_cm DECIMAL(8, 2) NOT NULL,
    width_cm DECIMAL(8, 2) NOT NULL,
    length_cm DECIMAL(8, 2) NOT NULL,
    weight_gms DECIMAL(8, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial default from user feedback
INSERT INTO package_matrix (product_count, height_cm, width_cm, length_cm, weight_gms)
VALUES (1, 10.00, 12.00, 19.00, 495.00)
ON CONFLICT (product_count) DO NOTHING;

-- Insert initial config placeholder (Actual token will be set via Settings UI)
INSERT INTO delhivery_configs (api_token, warehouse_name, env)
VALUES ('9de6fccd6360705681c74b6beed0829051064bce', 'Parfum Traders', 'staging')
ON CONFLICT DO NOTHING;
