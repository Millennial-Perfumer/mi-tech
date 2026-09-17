-- Migration 142: Remove the unused Azure Storage Account Key setting.
-- SMM Queue authentication uses the connection string, or an optional SAS
-- token with the account name.

DELETE FROM app_configs
WHERE key = 'azure_storage_account_key';
