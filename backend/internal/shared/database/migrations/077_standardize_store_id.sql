-- Standardize Store identifiers from legacy "1" to "shopify"

-- automation_triggers
UPDATE automation_triggers SET store_id = 'shopify' WHERE store_id = '1';

-- automation_messages
UPDATE automation_messages SET store_id = 'shopify' WHERE store_id = '1';

-- automation_templates
UPDATE automation_templates SET store_id = 'shopify' WHERE store_id = '1';

-- automation_whatsapp_settings
UPDATE automation_whatsapp_settings SET store_id = 'shopify' WHERE store_id = '1';

-- orders
UPDATE orders SET source_id = 'shopify' WHERE source_id = '1';
UPDATE orders SET store_id = 'shopify' WHERE store_id = '1';

-- customers
UPDATE customers SET source_id = 'shopify' WHERE source_id = '1';

-- webhook_events
UPDATE webhook_events SET source_id = 'shopify' WHERE source_id = '1';
