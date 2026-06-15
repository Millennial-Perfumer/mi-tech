-- Migration 110: B2B Payment Terms Table
CREATE TABLE IF NOT EXISTS b2b_payment_terms (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    due_days INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Seed default terms
INSERT INTO b2b_payment_terms (name, due_days) VALUES 
('Due on Receipt', 0),
('Net 15', 15),
('Net 30', 30),
('Net 45', 45),
('Net 60', 60)
ON CONFLICT (name) DO NOTHING;
