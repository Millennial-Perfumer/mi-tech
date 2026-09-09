-- Migration 138: Update the default UPI ID shown on B2B invoices.
-- Preserve a value deliberately configured by an administrator.
UPDATE app_configs
SET value = '7904769823@hdfc',
    label = 'UPI ID (For Payment QR)',
    category = 'business',
    sort_order = 9,
    updated_at = CURRENT_TIMESTAMP
WHERE key = 'upi_id'
  AND (BTRIM(value) = '' OR value = 'parfumtraders@upi');

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
SELECT
  'upi_id',
  '7904769823@hdfc',
  false,
  'UPI ID (For Payment QR)',
  'business',
  9
WHERE NOT EXISTS (
  SELECT 1 FROM app_configs WHERE key = 'upi_id'
);
