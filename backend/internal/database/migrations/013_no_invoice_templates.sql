-- Migration 013: Add no-invoice WhatsApp template variants

-- 1. order_placed_no_invoice_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id)
SELECT 
    store_id, 
    'order_placed_no_invoice_v3', 
    language, 
    category, 
    body, 
    '{"type": "none"}'::jsonb, 
    footer, 
    buttons, 
    'PENDING', 
    ''
FROM automation_templates 
WHERE template_name = 'order_placed_v3'
  AND NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_placed_no_invoice_v3');

-- 2. order_updated_no_invoice_v3
INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id)
SELECT 
    store_id, 
    'order_updated_no_invoice_v3', 
    language, 
    category, 
    body, 
    '{"type": "none"}'::jsonb, 
    footer, 
    buttons, 
    'PENDING', 
    ''
FROM automation_templates 
WHERE template_name = 'order_updated_v3'
  AND NOT EXISTS (SELECT 1 FROM automation_templates WHERE template_name = 'order_updated_no_invoice_v3');
