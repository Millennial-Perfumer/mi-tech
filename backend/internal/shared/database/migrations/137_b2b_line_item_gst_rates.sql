-- Store the selected GST rate on every B2B document line so v2 and v1
-- calculate the same tax after a document is saved and reopened.
ALTER TABLE b2b_invoice_items
    ADD COLUMN IF NOT EXISTS gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 18.00;

ALTER TABLE b2b_proforma_invoice_items
    ADD COLUMN IF NOT EXISTS gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 18.00;

ALTER TABLE b2b_credit_note_items
    ADD COLUMN IF NOT EXISTS gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 18.00;

ALTER TABLE b2b_debit_note_items
    ADD COLUMN IF NOT EXISTS gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 18.00;
