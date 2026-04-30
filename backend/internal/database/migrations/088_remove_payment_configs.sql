-- 088_remove_payment_configs.sql
-- Remove all Razorpay and payment related configurations since payment functionality is removed

DELETE FROM app_configs 
WHERE category = 'payment';
