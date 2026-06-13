-- 001_initial_schema.sql
-- Initial core schema for sources, orders, and line items

CREATE TABLE IF NOT EXISTS sources (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO sources (id, name) VALUES ('shopify', 'Shopify') ON CONFLICT (id) DO NOTHING;
INSERT INTO sources (id, name) VALUES ('amazon', 'Amazon') ON CONFLICT (id) DO NOTHING;
INSERT INTO sources (id, name) VALUES ('pos', 'POS') ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(255) PRIMARY KEY,
    source_id VARCHAR(50) NOT NULL DEFAULT 'shopify',
    external_order_id VARCHAR(255),
    order_number VARCHAR(255) NOT NULL,
    total_price DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    customer_name VARCHAR(255),
    customer_city VARCHAR(255),
    customer_state VARCHAR(100),
    customer_country VARCHAR(100),
    status VARCHAR(100),
    UNIQUE(source_id, external_order_id)
);

CREATE TABLE IF NOT EXISTS order_line_items (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) REFERENCES orders(id) ON DELETE CASCADE,
    title TEXT,
    sku VARCHAR(100),
    hs_code VARCHAR(50),
    quantity INTEGER,
    price DECIMAL(10, 2),
    discount DECIMAL(10, 2) DEFAULT 0,
    product_id VARCHAR(255),
    variant_id VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS webhook_events (
    id SERIAL PRIMARY KEY,
    source_id VARCHAR(50) NOT NULL,
    order_id VARCHAR(255) REFERENCES orders(id) ON DELETE SET NULL,
    topic VARCHAR(100) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    webhook_delivery_id VARCHAR(255) UNIQUE NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
