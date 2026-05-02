-- Migration: Add Flags for Inventory Deduction and Stock Addition
-- Created: 2026-05-02

ALTER TABLE manufacturing_oils ADD COLUMN deduct_inventory BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE manufacturing_products ADD COLUMN add_stock BOOLEAN NOT NULL DEFAULT TRUE;
