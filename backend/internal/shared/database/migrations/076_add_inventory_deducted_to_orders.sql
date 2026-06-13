-- Migration: 076_add_inventory_deducted_to_orders.sql
-- Description: Adds a flag to track if inventory has been deducted for an order.

ALTER TABLE orders ADD COLUMN IF NOT EXISTS inventory_deducted BOOLEAN DEFAULT FALSE;
