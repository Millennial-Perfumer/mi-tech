-- Migration 115: Add customer shipping address to invoices and proformas
ALTER TABLE b2b_invoices ADD COLUMN IF NOT EXISTS customer_shipping_address TEXT;
UPDATE b2b_invoices SET customer_shipping_address = customer_address WHERE customer_shipping_address IS NULL;

ALTER TABLE b2b_proforma_invoices ADD COLUMN IF NOT EXISTS customer_shipping_address TEXT;
UPDATE b2b_proforma_invoices SET customer_shipping_address = customer_address WHERE customer_shipping_address IS NULL;
