-- 002_order_updates.sql
-- Adding extended fields to orders and line items

ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS subtotal_price DECIMAL(12, 2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total_tax DECIMAL(12, 2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS store_id VARCHAR(255);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS currency VARCHAR(10);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS financial_status VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS fulfillment_status VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancel_reason TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS raw_payload JSONB;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_address1 TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_address2 TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_zip VARCHAR(20);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_first_name VARCHAR(255);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_last_name VARCHAR(255);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(100);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_company VARCHAR(100);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_url TEXT;

-- Migration to ensure constraints and foreign keys
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'orders_source_id_fkey') THEN
        ALTER TABLE orders ADD CONSTRAINT orders_source_id_fkey FOREIGN KEY (source_id) REFERENCES sources(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'webhook_events_source_id_fkey') THEN
        ALTER TABLE webhook_events ADD CONSTRAINT webhook_events_source_id_fkey FOREIGN KEY (source_id) REFERENCES sources(id);
    END IF;
END $$;
