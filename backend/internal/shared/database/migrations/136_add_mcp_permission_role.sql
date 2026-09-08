-- Store the reusable permission profile selected when an MCP machine key is created.
ALTER TABLE machine_api_keys
    ADD COLUMN IF NOT EXISTS permission_role TEXT NOT NULL DEFAULT 'custom';

ALTER TABLE machine_api_keys
    DROP CONSTRAINT IF EXISTS machine_api_keys_permission_role_check;

ALTER TABLE machine_api_keys
    ADD CONSTRAINT machine_api_keys_permission_role_check
    CHECK (permission_role IN ('read_only', 'full_access', 'custom'));

-- Preserve the reusable role for keys created before permission_role existed.
-- The catalog currently contains 37 scopes: 16 read, 13 write, and 8 destructive.
UPDATE machine_api_keys
SET permission_role = 'full_access'
WHERE permission_role = 'custom'
  AND cardinality(scopes) = 37
  AND scopes @> ARRAY[
      'orders:read', 'customers:read', 'metrics:read', 'gst:read',
      'inventory:read', 'production:read', 'b2b:read', 'communication:read',
      'marketing:read', 'feedback:read', 'abandoned_checkout:read', 'planner:read',
      'support:read', 'ai:read', 'settings:read', 'system:read',
      'orders:write', 'customers:write', 'inventory:write', 'production:write',
      'planner:write', 'b2b:write', 'communication:write', 'marketing:write',
      'feedback:write', 'support:write', 'settings:write', 'ai:write',
      'marketing:publish', 'orders:destructive', 'customers:destructive',
      'inventory:destructive', 'production:destructive', 'planner:destructive',
      'b2b:destructive', 'communication:destructive', 'ai:destructive'
  ]::text[];

UPDATE machine_api_keys
SET permission_role = 'read_only'
WHERE permission_role = 'custom'
  AND cardinality(scopes) = 16
  AND scopes @> ARRAY[
      'orders:read', 'customers:read', 'metrics:read', 'gst:read',
      'inventory:read', 'production:read', 'b2b:read', 'communication:read',
      'marketing:read', 'feedback:read', 'abandoned_checkout:read', 'planner:read',
      'support:read', 'ai:read', 'settings:read', 'system:read'
  ]::text[];
