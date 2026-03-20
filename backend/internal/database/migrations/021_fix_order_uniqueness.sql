-- 021_fix_order_uniqueness.sql
-- Update orders table to have a unique constraint on external_order_id only.
-- This prevents duplication when the same order is reported with different source names (e.g. 'amazon' vs 'shopify').

-- 1. Identify and remove any existing duplicates before applying the new constraint
-- (Optional but recommended for a clean migration)
-- DELETE FROM orders a USING orders b 
-- WHERE a.id > b.id AND a.external_order_id = b.external_order_id;

-- 2. Drop the old composite unique constraint
-- Note: The name of the constraint might vary depending on how it was created. 
-- In 001_initial_schema.sql it was created as UNIQUE(source_id, external_order_id).
-- PostgreSQL usually names this 'orders_source_id_external_order_id_key'.

DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'orders_source_id_external_order_id_key') THEN
        ALTER TABLE orders DROP CONSTRAINT orders_source_id_external_order_id_key;
    END IF;
END $$;

-- 3. Add the new unique constraint on external_order_id only
ALTER TABLE orders ADD CONSTRAINT orders_external_order_id_key UNIQUE (external_order_id);
