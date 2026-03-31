DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='order_line_items' AND column_name='order_discount') THEN
        ALTER TABLE order_line_items ADD COLUMN order_discount NUMERIC(10,2) DEFAULT 0;
    END IF;
END $$;
