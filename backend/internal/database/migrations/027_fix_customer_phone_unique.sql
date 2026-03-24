-- Drop the existing unique constraint if it exists (it's usually named after the table and column)
-- PostgreSQL standard name for UNIQUE(phone_number) on table customers
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_phone_number_key;

-- Create a unique index that only applies to non-deleted records
-- This allows reusing a phone number if the previous record was soft-deleted
CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_phone_unique_active 
ON customers(phone_number) 
WHERE deleted_at IS NULL;
