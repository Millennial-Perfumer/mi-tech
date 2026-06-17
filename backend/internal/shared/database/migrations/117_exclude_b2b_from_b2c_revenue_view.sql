-- Migration 117: Exclude B2B from B2C orders in unified view

-- Register B2B source
INSERT INTO sources (id, name, enabled) VALUES ('b2b', 'B2B', true) ON CONFLICT (id) DO NOTHING;

-- Backfill existing B2B Invoices to orders table
INSERT INTO orders (
    source_id, 
    external_order_id, 
    order_number, 
    invoice_number, 
    total_price, 
    subtotal_price, 
    total_tax, 
    currency, 
    financial_status, 
    fulfillment_status, 
    status, 
    customer_name, 
    customer_phone, 
    customer_email, 
    customer_address1, 
    customer_state, 
    total_discount, 
    created_at, 
    updated_at
)
SELECT 
    'b2b', 
    'B2B-' || id::varchar, 
    COALESCE(invoice_number, 'B2B-' || id::varchar), 
    invoice_number, 
    total_price, 
    subtotal_price, 
    (cgst_amount + sgst_amount + igst_amount), 
    'INR', 
    CASE 
        WHEN LOWER(payment_status) = 'paid' THEN 'paid'
        WHEN LOWER(payment_status) = 'partial' THEN 'partially_paid'
        ELSE 'unpaid'
    END, 
    'fulfilled', 
    LOWER(status), 
    customer_name, 
    customer_phone, 
    customer_email, 
    customer_address, 
    customer_state, 
    discount_amount, 
    invoice_date, 
    updated_at
FROM b2b_invoices
ON CONFLICT (external_order_id) DO UPDATE SET 
    order_number = EXCLUDED.order_number,
    invoice_number = EXCLUDED.invoice_number,
    total_price = EXCLUDED.total_price,
    subtotal_price = EXCLUDED.subtotal_price,
    total_tax = EXCLUDED.total_tax,
    financial_status = EXCLUDED.financial_status,
    status = EXCLUDED.status,
    customer_name = EXCLUDED.customer_name,
    customer_phone = EXCLUDED.customer_phone,
    customer_email = EXCLUDED.customer_email,
    customer_address1 = EXCLUDED.customer_address1,
    customer_state = EXCLUDED.customer_state,
    total_discount = EXCLUDED.total_discount,
    updated_at = EXCLUDED.updated_at;

-- Backfill existing B2B Invoice items to order_line_items table
INSERT INTO order_line_items (
    id, 
    order_id, 
    product_id, 
    title, 
    sku, 
    hs_code, 
    quantity, 
    price, 
    discount, 
    order_discount
)
SELECT 
    'b2b-' || ii.invoice_id::varchar || '-' || ii.id::varchar, 
    o.id, 
    ii.product_id::varchar, 
    ii.item_details, 
    ii.sku, 
    ii.hsn_code, 
    ii.quantity::int, 
    ii.rate, 
    0.0, 
    0.0
FROM b2b_invoice_items ii
JOIN orders o ON o.external_order_id = 'B2B-' || ii.invoice_id::varchar
ON CONFLICT (id) DO UPDATE SET
    quantity = EXCLUDED.quantity,
    price = EXCLUDED.price,
    title = EXCLUDED.title,
    sku = EXCLUDED.sku,
    hs_code = EXCLUDED.hs_code;

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
