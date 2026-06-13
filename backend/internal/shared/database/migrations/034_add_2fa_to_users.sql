-- Migration: 034_add_2fa_to_users
-- Description: Add 2FA fields to users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_code VARCHAR(10);
ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_expiry TIMESTAMP WITH TIME ZONE;

-- Remark: We default two_factor_enabled to TRUE for security.
-- Note: Existing users without phone numbers will need an admin to add them before they can login.
