-- Migration 141: Make SMM Queue Azure storage configurable from Settings.
-- Secret values are masked in the settings UI and are never returned by the
-- normal configs endpoint in plaintext.

INSERT INTO app_configs (key, value, is_secret, label, category, sort_order)
VALUES
    ('azure_storage_connection_string', '', true, 'Azure Storage Connection String', 'smm_queue', 1),
    ('azure_storage_account_name', 'mptechstg', false, 'Azure Storage Account Name', 'smm_queue', 2),
    ('azure_storage_sas_token', '', true, 'Azure Storage SAS Token (Optional)', 'smm_queue', 3),
    ('smm_queue_container', 'mp-smm-queue', false, 'SMM Queue Container', 'smm_queue', 4)
ON CONFLICT (key) DO UPDATE SET
    is_secret = EXCLUDED.is_secret,
    label = EXCLUDED.label,
    category = EXCLUDED.category,
    sort_order = EXCLUDED.sort_order;
