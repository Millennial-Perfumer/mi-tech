-- Migration: 101_increase_inventory_title_length
-- Description: Alter inventory_items title column to TEXT to handle long product and variant names.

ALTER TABLE inventory_items ALTER COLUMN title TYPE TEXT;
