-- Migration 112: B2B Proforma Invoices and Conversions

-- B2B Proforma Invoices Table
CREATE TABLE IF NOT EXISTS b2b_proforma_invoices (
    id BIGSERIAL PRIMARY KEY,
    proforma_number VARCHAR(100),
    proforma_sequence INT,
    financial_year VARCHAR(10),
    note_date DATE NOT NULL,
    valid_until DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'DRAFT', -- 'DRAFT', 'SENT', 'ACCEPTED', 'CONVERTED_TO_INVOICE', 'REJECTED', 'EXPIRED', 'CANCELLED'
    revision_number INT DEFAULT 1,
    parent_proforma_id BIGINT,

    -- Customer snapshot (historical details)
    customer_id BIGINT REFERENCES b2b_customers(id) ON DELETE SET NULL,
    customer_gstin VARCHAR(15) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    customer_state VARCHAR(100) NOT NULL,
    customer_state_code VARCHAR(2) NOT NULL,
    customer_address TEXT NOT NULL,
    
    -- Seller Details snapshot
    seller_gstin VARCHAR(15) NOT NULL,
    seller_name VARCHAR(255) NOT NULL,
    seller_state VARCHAR(100) NOT NULL,
    seller_state_code VARCHAR(2) NOT NULL,
    seller_address TEXT NOT NULL,

    -- Financial pricing summaries
    subtotal_price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    discount_percent NUMERIC(5, 2) DEFAULT 0.00,
    discount_amount NUMERIC(15, 2) DEFAULT 0.00,
    
    -- GST split details
    cgst_rate NUMERIC(5, 2) DEFAULT 0.00,
    cgst_amount NUMERIC(15, 2) DEFAULT 0.00,
    sgst_rate NUMERIC(5, 2) DEFAULT 0.00,
    sgst_amount NUMERIC(15, 2) DEFAULT 0.00,
    igst_rate NUMERIC(5, 2) DEFAULT 0.00,
    igst_amount NUMERIC(15, 2) DEFAULT 0.00,
    
    -- Final Totals
    total_price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    advance_paid NUMERIC(15, 2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_b2b_proforma_date ON b2b_proforma_invoices(note_date);

-- B2B Proforma Invoice Items Table
CREATE TABLE IF NOT EXISTS b2b_proforma_invoice_items (
    id BIGSERIAL PRIMARY KEY,
    proforma_invoice_id BIGINT REFERENCES b2b_proforma_invoices(id) ON DELETE CASCADE,
    product_id BIGINT,
    item_details VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    hsn_code VARCHAR(8),
    quantity NUMERIC(15, 4) NOT NULL DEFAULT 1.0000,
    rate NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00
);

-- Alter B2B Invoices Table to link to Proforma
ALTER TABLE b2b_invoices ADD COLUMN IF NOT EXISTS proforma_id BIGINT REFERENCES b2b_proforma_invoices(id) ON DELETE SET NULL;
ALTER TABLE b2b_invoices ADD COLUMN IF NOT EXISTS advance_adjusted NUMERIC(15, 2) DEFAULT 0.00;
