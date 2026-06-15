-- 105_add_invoice_number_to_orders.sql
-- Add invoice_number column, backfill existing orders, and create triggers for auto-sequencing.

ALTER TABLE orders ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(255);

CREATE TABLE IF NOT EXISTS invoice_sequences (
    source_id VARCHAR(50) PRIMARY KEY,
    current_value BIGINT NOT NULL DEFAULT 0
);

WITH RankedOrders AS (
    SELECT 
        id,
        source_id,
        order_number,
        ROW_NUMBER() OVER (PARTITION BY source_id ORDER BY created_at ASC) as rn
    FROM orders
)
UPDATE orders o
SET invoice_number = CASE 
    WHEN ro.source_id = 'amazon' THEN 'AMZ-' || ro.rn
    WHEN ro.source_id = 'shopify' THEN 'SY-' || COALESCE(NULLIF(regexp_replace(ro.order_number, '[^0-9]', '', 'g'), ''), ro.rn::text)
    ELSE UPPER(ro.source_id) || '-' || ro.rn
END
FROM RankedOrders ro
WHERE o.id = ro.id;

INSERT INTO invoice_sequences (source_id, current_value)
SELECT source_id, COUNT(*)
FROM orders
GROUP BY source_id
ON CONFLICT (source_id) DO UPDATE 
SET current_value = EXCLUDED.current_value;

CREATE OR REPLACE FUNCTION set_order_invoice_number()
RETURNS TRIGGER AS $$
DECLARE
    next_val BIGINT;
    prefix VARCHAR(50);
    digits TEXT;
BEGIN
    IF NEW.invoice_number IS NULL OR NEW.invoice_number = '' THEN
        IF NEW.source_id = 'shopify' THEN
            digits := NULLIF(regexp_replace(NEW.order_number, '[^0-9]', '', 'g'), '');
            IF digits IS NOT NULL THEN
                NEW.invoice_number := 'SY-' || digits;
                RETURN NEW;
            END IF;
            prefix := 'SY-';
        ELSIF NEW.source_id = 'amazon' THEN
            prefix := 'AMZ-';
        ELSIF NEW.source_id = 'pos' THEN
            prefix := 'POS-';
        ELSE
            prefix := UPPER(NEW.source_id) || '-';
        END IF;

        INSERT INTO invoice_sequences (source_id, current_value)
        VALUES (NEW.source_id, 1)
        ON CONFLICT (source_id)
        DO UPDATE SET current_value = invoice_sequences.current_value + 1
        RETURNING current_value INTO next_val;

        NEW.invoice_number := prefix || next_val;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_set_order_invoice_number ON orders;

CREATE TRIGGER trigger_set_order_invoice_number
BEFORE INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION set_order_invoice_number();
