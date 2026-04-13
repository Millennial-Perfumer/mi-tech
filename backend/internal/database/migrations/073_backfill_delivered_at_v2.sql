-- 073_backfill_delivered_at_v2.sql
-- One-time script to backfill orders.delivered_at for delivered orders with phones
-- This handles legacy data that missed the timestamp during status transitions.

UPDATE orders 
SET delivered_at = NOW(),
    feedback_status_id = COALESCE(feedback_status_id, 1),
    updated_at = NOW()
WHERE customer_phone IS NOT NULL 
  AND customer_phone != '' 
  AND delivered_at IS NULL 
  AND (TRIM(LOWER(delivery_status)) = 'delivered' OR TRIM(LOWER(status)) = 'delivered');
