-- Migration 108: Add notes column to b2b_customers table
ALTER TABLE b2b_customers ADD COLUMN IF NOT EXISTS notes TEXT;
