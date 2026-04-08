-- 061_revert_flow_configs.sql
-- Remove WhatsApp Flow private key from app_configs

DELETE FROM app_configs WHERE key = 'whatsapp_flow_private_key';
