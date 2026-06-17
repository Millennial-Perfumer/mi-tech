-- Migration 114: Remove 'Clear All' button config from settings
DELETE FROM app_configs WHERE key = 'show_clear_customers_button';
DELETE FROM app_settings WHERE key = 'show_clear_customers_button';
