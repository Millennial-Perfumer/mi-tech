-- Migration 111: B2B Credit/Debit Notes, Locks, Ledger & Audit Trail

-- B2B Credit Notes Table
CREATE TABLE IF NOT EXISTS b2b_credit_notes (
    id BIGSERIAL PRIMARY KEY,
    credit_note_number VARCHAR(100) UNIQUE,
    credit_note_sequence INT,
    financial_year VARCHAR(10),
    invoice_id BIGINT REFERENCES b2b_invoices(id) ON DELETE SET NULL,
    invoice_number VARCHAR(100),
    note_date DATE NOT NULL,
    reason VARCHAR(255),
    
    -- Customer snapshot (captures historical details for the credit note)
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
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- 'DRAFT', 'ISSUED', 'CANCELLED'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_b2b_credit_notes_date ON b2b_credit_notes(note_date);

-- B2B Credit Note Items Table
CREATE TABLE IF NOT EXISTS b2b_credit_note_items (
    id BIGSERIAL PRIMARY KEY,
    credit_note_id BIGINT REFERENCES b2b_credit_notes(id) ON DELETE CASCADE,
    product_id BIGINT,
    item_details VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    hsn_code VARCHAR(8),
    quantity NUMERIC(15, 4) NOT NULL DEFAULT 1.0000,
    rate NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00
);

-- B2B Debit Notes Table
CREATE TABLE IF NOT EXISTS b2b_debit_notes (
    id BIGSERIAL PRIMARY KEY,
    debit_note_number VARCHAR(100) UNIQUE,
    debit_note_sequence INT,
    financial_year VARCHAR(10),
    invoice_id BIGINT REFERENCES b2b_invoices(id) ON DELETE SET NULL,
    invoice_number VARCHAR(100),
    note_date DATE NOT NULL,
    reason VARCHAR(255),
    
    -- Customer snapshot
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
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- 'DRAFT', 'ISSUED', 'CANCELLED'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_b2b_debit_notes_date ON b2b_debit_notes(note_date);

-- B2B Debit Note Items Table
CREATE TABLE IF NOT EXISTS b2b_debit_note_items (
    id BIGSERIAL PRIMARY KEY,
    debit_note_id BIGINT REFERENCES b2b_debit_notes(id) ON DELETE CASCADE,
    product_id BIGINT,
    item_details VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    hsn_code VARCHAR(8),
    quantity NUMERIC(15, 4) NOT NULL DEFAULT 1.0000,
    rate NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00
);

-- GST Period Locks Table
CREATE TABLE IF NOT EXISTS gst_periods (
    id BIGSERIAL PRIMARY KEY,
    month INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN', -- 'OPEN', 'LOCKED'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_gst_periods UNIQUE (month, year)
);

-- B2B Financial Audit Logs Table
CREATE TABLE IF NOT EXISTS b2b_financial_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL, -- 'INVOICE', 'CREDIT_NOTE', 'DEBIT_NOTE', 'PAYMENT'
    entity_id BIGINT NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    description TEXT,
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Recreate unified view to include Credit Notes and Debit Notes
DROP VIEW IF EXISTS unified_revenue_transactions;
CREATE VIEW unified_revenue_transactions AS
SELECT 
    id::varchar as transaction_id,
    created_at as tx_date,
    total_price,
    total_discount,
    customer_state as state,
    'B2C_ORDER' as source_type,
    status as order_status,
    fulfillment_status as fulfillment_status,
    financial_status as payment_status,
    source_id
FROM orders

UNION ALL

SELECT 
    'B2B-' || id::varchar as transaction_id,
    created_at as tx_date,
    total_price,
    discount_amount as total_discount,
    customer_state as state,
    'B2B_INVOICE' as source_type,
    status as order_status,
    'fulfilled' as fulfillment_status,
    payment_status as payment_status,
    'B2B' as source_id
FROM b2b_invoices
WHERE status = 'ISSUED'

UNION ALL

SELECT 
    'B2B-CN-' || id::varchar as transaction_id,
    created_at as tx_date,
    -total_price as total_price,
    -discount_amount as total_discount,
    customer_state as state,
    'B2B_CREDIT_NOTE' as source_type,
    status as order_status,
    'fulfilled' as fulfillment_status,
    'paid' as payment_status,
    'B2B' as source_id
FROM b2b_credit_notes
WHERE status = 'ISSUED'

UNION ALL

SELECT 
    'B2B-DN-' || id::varchar as transaction_id,
    created_at as tx_date,
    total_price,
    discount_amount as total_discount,
    customer_state as state,
    'B2B_DEBIT_NOTE' as source_type,
    status as order_status,
    'fulfilled' as fulfillment_status,
    'unpaid' as payment_status,
    'B2B' as source_id
FROM b2b_debit_notes
WHERE status = 'ISSUED';
