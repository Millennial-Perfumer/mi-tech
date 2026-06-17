-- Migration 116: Add inventory_deducted flag to B2B Invoices
ALTER TABLE b2b_invoices ADD COLUMN IF NOT EXISTS inventory_deducted BOOLEAN DEFAULT FALSE;
