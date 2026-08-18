-- 124_resequence_amazon_invoices.sql
-- 1. Resequence existing Amazon orders chronologically by created_at
WITH RankedAmazonOrders AS (
    SELECT 
        id,
        ROW_NUMBER() OVER (ORDER BY created_at ASC) as rn
    FROM orders
    WHERE source_id = 'amazon'
)
UPDATE orders o
SET invoice_number = 'AMZ-' || rao.rn
FROM RankedAmazonOrders rao
WHERE o.id = rao.id;

-- 2. Update/Reset the sequence tracker to match actual count
INSERT INTO invoice_sequences (source_id, current_value)
VALUES ('amazon', (SELECT COALESCE(COUNT(*), 0) FROM orders WHERE source_id = 'amazon'))
ON CONFLICT (source_id) DO UPDATE 
SET current_value = EXCLUDED.current_value;

-- 3. Fix the trigger function to prevent sequence increment on GORM Upsert conflict (ON CONFLICT DO UPDATE)
CREATE OR REPLACE FUNCTION set_order_invoice_number()
RETURNS TRIGGER AS $$
DECLARE
    next_val BIGINT;
    prefix VARCHAR(50);
    digits TEXT;
    existing_inv VARCHAR(255);
BEGIN
    -- If the order already exists in the database, preserve its invoice number
    -- to prevent sequence leaks during INSERT ... ON CONFLICT DO UPDATE (Upsert)
    SELECT invoice_number INTO existing_inv 
    FROM orders 
    WHERE source_id = NEW.source_id AND external_order_id = NEW.external_order_id;

    IF existing_inv IS NOT NULL AND existing_inv != '' THEN
        NEW.invoice_number := existing_inv;
        RETURN NEW;
    END IF;

    -- Sequence generation for new orders
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
