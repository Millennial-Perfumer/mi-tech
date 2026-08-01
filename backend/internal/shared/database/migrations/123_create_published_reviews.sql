-- Migration: Create Published Reviews table for Judge.me tracking
CREATE TABLE IF NOT EXISTS published_reviews (
    id SERIAL PRIMARY KEY,
    review_id VARCHAR(100),
    product_id VARCHAR(100) NOT NULL,
    product_title VARCHAR(255) NOT NULL,
    reviewer_name VARCHAR(150) NOT NULL,
    gender VARCHAR(20) DEFAULT 'unspecified',
    email VARCHAR(150) NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    shop_domain VARCHAR(150) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'SUCCESS',
    status_code INT NOT NULL DEFAULT 200,
    published_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_published_reviews_product_id ON published_reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_published_reviews_published_at ON published_reviews(published_at DESC);
