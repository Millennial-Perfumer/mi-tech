-- 068_backfill_delivered_at.sql
-- One-time script to backfill orders.delivered_at from WhatsApp message logs

UPDATE orders o
SET delivered_at = subquery.earliest_sent
FROM (
    SELECT 
        m.order_id, 
        MIN(m.sent_at) as earliest_sent
    FROM automation_messages m
    JOIN automation_templates t ON m.template_id = t.id
    WHERE t.template_name = 'order_delivered_v3_1'
    GROUP BY m.order_id
) AS subquery
WHERE o.id = subquery.order_id
  AND o.delivered_at IS NULL;
