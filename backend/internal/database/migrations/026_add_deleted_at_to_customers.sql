-- Add deleted_at column for soft deletes in customers table
ALTER TABLE customers ADD COLUMN deleted_at TIMESTAMP;
CREATE INDEX idx_customers_deleted_at ON customers(deleted_at);
