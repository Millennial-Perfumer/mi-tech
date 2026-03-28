-- Migration: 033_cleanup_external_users
-- Description: Remove users whose usernames (emails) do not belong to @millennialperfumer.in

DELETE FROM users 
WHERE username NOT LIKE '%@millennialperfumer.in';
