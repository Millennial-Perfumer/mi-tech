-- 018_convert_automation_order_id_to_bigint.sql

BEGIN;

-- 1. Create temporary column
ALTER TABLE automation_messages ADD COLUMN order_id_new BIGINT;

-- 2. Convert data (handling existing string IDs)
UPDATE automation_messages SET order_id_new = order_id::BIGINT;

-- 3. Drop old column and rename
ALTER TABLE automation_messages DROP COLUMN order_id;
ALTER TABLE automation_messages RENAME COLUMN order_id_new TO order_id;

-- 4. Set NOT NULL
ALTER TABLE automation_messages ALTER COLUMN order_id SET NOT NULL;

COMMIT;
