-- Migration 120: Add Abandoned Checkout Delay Configuration
-- This allows setting the delay in minutes before an abandoned checkout message is recovery triggered.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order, updated_at)
VALUES (
    'abandoned_checkout_delay_minutes',
    '30',
    false,
    'Abandoned Checkout Recovery Delay (Minutes)',
    'abandoned_cart',
    90,
    NOW()
)
ON CONFLICT (key) DO UPDATE SET 
    label = EXCLUDED.label,
    category = EXCLUDED.category,
    sort_order = EXCLUDED.sort_order;
