-- 007_v3_template_migration.sql
-- Single Source of Truth for V3 WhatsApp Templates
-- All templates unified to Category: UTILITY with Body-Link dynamic tracking

-- 1. Seeding V3 Templates (Idempotent)
-- order_placed_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'order_placed_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} has been confirmed.

Please find your invoice attached. Tracking details will be shared once your order is shipped.

— Millennial Perfumer',
'{"type": "DOCUMENT", "sample": "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"}'::jsonb,
'Crafted for Impact',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_placed_v3');

-- order_assigned_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'order_assigned_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} has been assigned to {{3}}.

Tracking ID: {{4}}
Track here: {{5}}

— Millennial Perfumer',
'{"type": "none", "sample": "Aboobaker, 2429, Delhivery, 123456, https://tracking.com"}'::jsonb,
'Every scent tells a story.',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_assigned_v3');

-- order_dispatched_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'order_dispatched_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} has been dispatched via {{3}}.
Tracking ID: {{4}}
Track here: {{5}}

— Millennial Perfumer',
'{"type": "none", "sample": "Aboobaker, 2429, Delhivery, 123456, https://tracking.com"}'::jsonb,
'Not just perfume. A memory.',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_dispatched_v3');

-- out_for_delivery_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'out_for_delivery_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} is out for delivery!
Courier: {{3}} ({{4}})
Track here: {{5}}

— Millennial Perfumer',
'{"type": "none", "sample": "Aboobaker, 2429, Delhivery, 123456, https://tracking.com"}'::jsonb,
'Crafted for Impact',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'out_for_delivery_v3');

-- order_delivered_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'order_delivered_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} has been delivered.

We hope you enjoy your fragrance.

— Millennial Perfumer',
'{"type": "none", "sample": "Aboobaker, 2429"}'::jsonb,
'Crafted in scent. Remembered forever.',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_delivered_v3');

-- order_cancelled_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'order_cancelled_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} has been cancelled.

If a payment was made, the refund will be processed shortly.

— Millennial Perfumer',
'{"type": "none", "sample": "Aboobaker, 2429"}'::jsonb,
'Crafted for Impact',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_cancelled_v3');

-- order_updated_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status)
SELECT '1', 'order_updated_v3', 'en_US', 'UTILITY', 
'Hi {{1}},

Your order #{{2}} has been updated.

Please find the updated invoice attached.

— Millennial Perfumer',
'{"type": "DOCUMENT", "sample": "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"}'::jsonb,
'Crafted for Impact',
'[]'::jsonb,
'PENDING'
WHERE NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_updated_v3');

-- 2. Seeding Triggers for V3 Templates (Idempotent)
INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/create', id, true FROM automation_templates WHERE template_name = 'order_placed_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/create' AND store_id = '1');

INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/assigned', id, true FROM automation_templates WHERE template_name = 'order_assigned_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/assigned' AND store_id = '1');

INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/fulfilled', id, true FROM automation_templates WHERE template_name = 'order_dispatched_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/fulfilled' AND store_id = '1');

INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/out_for_delivery', id, true FROM automation_templates WHERE template_name = 'out_for_delivery_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/out_for_delivery' AND store_id = '1');

INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/delivered', id, true FROM automation_templates WHERE template_name = 'order_delivered_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/delivered' AND store_id = '1');

INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/cancelled', id, true FROM automation_templates WHERE template_name = 'order_cancelled_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/cancelled' AND store_id = '1');

INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
SELECT '1', 'orders/updated', id, true FROM automation_templates WHERE template_name = 'order_updated_v3'
AND NOT EXISTS (SELECT 1 FROM automation_triggers WHERE webhook_topic = 'orders/updated' AND store_id = '1');
