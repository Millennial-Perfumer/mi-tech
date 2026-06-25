-- 122_add_location_to_abandoned_checkouts.sql
-- Add location/address details to abandoned checkouts for analytics and custom routing

ALTER TABLE abandoned_checkouts 
ADD COLUMN IF NOT EXISTS city TEXT,
ADD COLUMN IF NOT EXISTS province TEXT,
ADD COLUMN IF NOT EXISTS country TEXT,
ADD COLUMN IF NOT EXISTS zip TEXT;
