-- Migration 108: Add gst_rate column to b2b_invoice_items
-- This stores the GST rate selected per line item so it survives edit/reload cycles

ALTER TABLE b2b_invoice_items
    ADD COLUMN IF NOT EXISTS gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 18.00;
