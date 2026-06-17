-- Migration 107: B2B Invoicing & Customers System

-- B2B Customers Table
CREATE TABLE IF NOT EXISTS b2b_customers (
    id BIGSERIAL PRIMARY KEY,
    legal_name VARCHAR(255) NOT NULL,
    trade_name VARCHAR(255),
    gstin VARCHAR(15) UNIQUE NOT NULL,
    pan VARCHAR(10),
    email VARCHAR(255),
    phone VARCHAR(50),
    billing_address TEXT NOT NULL,
    shipping_address TEXT,
    state VARCHAR(100) NOT NULL,
    state_code VARCHAR(2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for lookup via GSTIN
CREATE INDEX IF NOT EXISTS idx_b2b_customers_gstin ON b2b_customers(gstin);

-- B2B Invoices Table
CREATE TABLE IF NOT EXISTS b2b_invoices (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(100) UNIQUE,
    invoice_sequence INT,
    financial_year VARCHAR(10),
    order_number VARCHAR(100),
    invoice_date DATE NOT NULL,
    terms VARCHAR(100),
    due_date DATE,
    salesperson VARCHAR(255),
    subject VARCHAR(500),
    
    -- Customer snapshot (captures historical details for the invoice)
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
    
    -- Additional Charges
    tds_tcs_type VARCHAR(20) DEFAULT 'NONE', -- 'TDS', 'TCS', 'NONE'
    tds_tcs_rate NUMERIC(5, 2) DEFAULT 0.00,
    tds_tcs_amount NUMERIC(15, 2) DEFAULT 0.00,
    transportation_charge NUMERIC(15, 2) DEFAULT 0.00,
    
    -- Final Totals
    total_price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    
    -- Payment and lifecycle statuses
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- 'DRAFT', 'ISSUED', 'CANCELLED'
    payment_status VARCHAR(20) NOT NULL DEFAULT 'UNPAID', -- 'UNPAID', 'PARTIAL', 'PAID'
    paid_amount NUMERIC(15, 2) DEFAULT 0.00,
    balance_amount NUMERIC(15, 2) DEFAULT 0.00,
    payment_date DATE,
    payment_method VARCHAR(50),

    customer_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for quick sorting & lookup
CREATE INDEX IF NOT EXISTS idx_b2b_invoices_date ON b2b_invoices(invoice_date);

-- B2B Invoice Items Table
CREATE TABLE IF NOT EXISTS b2b_invoice_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT REFERENCES b2b_invoices(id) ON DELETE CASCADE,
    product_id BIGINT,
    item_details VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    hsn_code VARCHAR(8),
    quantity NUMERIC(15, 4) NOT NULL DEFAULT 1.0000,
    rate NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00
);

-- Unified views for Dashboard Metrics & Reporting
CREATE OR REPLACE VIEW unified_revenue_transactions AS
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
WHERE status = 'ISSUED';
