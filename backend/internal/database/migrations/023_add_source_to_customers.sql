-- Migration: 023_add_source_to_customers.sql
-- Add source_id column to customers table for tracking origins.

ALTER TABLE customers ADD COLUMN IF NOT EXISTS source_id VARCHAR(50);

-- Ensure default sources exist and add "manual"
INSERT INTO sources (id, name, enabled, created_at)
VALUES 
    ('shopify', 'Shopify Store', true, CURRENT_TIMESTAMP),
    ('amazon', 'Amazon India', true, CURRENT_TIMESTAMP),
    ('pos', 'Retail POS', true, CURRENT_TIMESTAMP),
    ('manual', 'Manual CSV Import', true, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;
