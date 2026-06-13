-- 067_fix_feedback_schema.sql
-- Add missing updated_at column and unique constraint for order_id

-- 1. Add updated_at column
ALTER TABLE customer_feedback ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- 2. Add Unique constraint on order_id to allow upsert (OnConflict)
-- We check if it exists first to avoid errors. 
-- In Postgres, we can do this by using a DO block or just adding it (index or constraint).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'customer_feedback_order_id_key'
    ) THEN
        ALTER TABLE customer_feedback ADD CONSTRAINT customer_feedback_order_id_key UNIQUE (order_id);
    END IF;
END $$;
