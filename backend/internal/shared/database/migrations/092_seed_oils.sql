-- Migration: Seed Oil Inventory and Link to Products
-- Created: 2026-05-01

-- Add unique constraint to name if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'oil_inventory_name_unique') THEN
        ALTER TABLE oil_inventory ADD CONSTRAINT oil_inventory_name_unique UNIQUE (name);
    END IF;
END $$;

INSERT INTO oil_inventory (name, inventory_item_id)
SELECT data.name, items.id
FROM (
    VALUES 
        ('Ocean Drift', 'mi-001'), ('Floral Essence', 'mi-002'), ('Dark Dreams', 'mi-003'),
        ('Desert King', 'mi-004'), ('Oudh Royale', 'mi-005'), ('Imagine', 'mi-006'),
        ('Orchid', 'mi-009'), ('Urban Pulse', 'mi-010'), ('Reverie', 'mi-011'),
        ('Pinnacle', 'mi-012'), ('Surge', 'mi-013'), ('Velocity', 'mi-014'),
        ('One', 'mi-015'), ('Guilt', 'mi-016'), ('Millionaire', 'mi-017'),
        ('Instinct', 'mi-018'), ('Terra Noir', 'mi-019'), ('Leather', 'mi-020'),
        ('Allure', 'mi-021'), ('Silver', 'mi-022'), ('Nomade', 'mi-023'),
        ('Fabulous', 'mi-024'), ('Vanille', 'mi-025'), ('Stronger', 'mi-026'),
        ('Pegasus', 'mi-027'), ('Chill', 'mi-028'), ('Aqua', 'mi-029'),
        ('Code', 'mi-030'), ('Male', 'mi-031'), ('Y', 'mi-032'),
        ('Maracuja', 'mi-033'), ('Pride Men', 'mi-034'), ('Tweed', 'mi-035'),
        ('Khamrah', 'mi-036'), ('Sage', 'mi-039'), ('Interlude', 'mi-040'),
        ('Intense Men', 'mi-041'), ('Arcade', 'mi-042'), ('Vanilla Day', 'mi-043'),
        ('No. 5', 'mi-044'), ('Cherry', 'mi-045'), ('Angel Share', 'mi-046'),
        ('Imperial', 'mi-047'), ('Aeros', 'mi-048'), ('Spicebomb Ext', 'mi-049'),
        ('Tuxedo', 'mi-050'), ('Her', 'mi-051'), ('Oud Satin Mood', 'mi-052'),
        ('All Of Me', 'mi-053'), ('VIP Men', 'mi-054'), ('Rose Oudh', 'mi-055'),
        ('Oud Intense', 'mi-056'), ('Midday Swim', 'mi-057'), ('Pure Aura', 'mi-058'),
        ('Rome Intense', 'mi-059'), ('Viking', 'mi-060'), ('Fireplace', 'mi-061'),
        ('Layton', 'mi-062'), ('Myself', 'mi-063'), ('Victory', 'mi-064'),
        ('Dylan', 'mi-065'), ('Coco', 'mi-066'), ('Male Elixir', 'mi-067'),
        ('Guidance', 'mi-068'), ('Talisman Blue', 'mi-069'), ('Symphony', 'mi-070'),
        ('Vanira', 'mi-071'), ('Tygar', 'mi-072')
) AS data(name, sku)
JOIN inventory_items items ON items.mi_sku = data.sku
ON CONFLICT (name) DO NOTHING;
