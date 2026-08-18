-- Migration 129: Add origin_order_id to b2b_invoices
ALTER TABLE b2b_invoices ADD COLUMN IF NOT EXISTS origin_order_id VARCHAR(255);
