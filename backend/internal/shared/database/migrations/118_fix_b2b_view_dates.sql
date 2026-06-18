-- Migration 118: Use transaction date instead of created_at for B2B invoices, credit notes, and debit notes in unified_revenue_transactions view

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
WHERE COALESCE(source_id, '') != 'b2b'

UNION ALL

SELECT 
    'B2B-' || id::varchar as transaction_id,
    invoice_date as tx_date,
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
    note_date as tx_date,
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
    note_date as tx_date,
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
