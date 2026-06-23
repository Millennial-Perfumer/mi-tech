-- Migration 121: Update Abandoned Checkout Delay Category
-- This moves the abandoned checkout delay key to the new 'abandoned_cart' category.

UPDATE app_configs 
SET category = 'abandoned_cart' 
WHERE key = 'abandoned_checkout_delay_minutes';
