package db

import (
	"database/sql"
	"fmt"
	"log"

	"shopify-gst-app/internal/config"

	_ "github.com/lib/pq"
)

// InitDB connects to the PostgreSQL database and returns the connection pool.
// It also runs auto-migrations to ensure tables exist.
func InitDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	// Run migrations
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	return db, nil
}

// migrate creates the required tables if they do not exist
func migrate(db *sql.DB) error {
	// 1. Sources Table (Dependencies should be created first)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sources (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO sources (id, name) VALUES ('shopify', 'Shopify') ON CONFLICT (id) DO NOTHING;
		INSERT INTO sources (id, name) VALUES ('amazon', 'Amazon') ON CONFLICT (id) DO NOTHING;
		INSERT INTO sources (id, name) VALUES ('pos', 'POS') ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		return fmt.Errorf("failed to create sources table: %w", err)
	}

	// 2. Orders Table
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(255) PRIMARY KEY,
		source_id VARCHAR(50) NOT NULL DEFAULT 'shopify',
		external_order_id VARCHAR(255),
		order_number VARCHAR(255) NOT NULL,
		total_price DECIMAL(12, 2) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		customer_name VARCHAR(255),
		customer_city VARCHAR(255),
		customer_state VARCHAR(100),
		customer_country VARCHAR(100),
		status VARCHAR(100),
		UNIQUE(source_id, external_order_id)
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	// Gracefully add columns to existing tables (will error innocuously if already exists, so we ignore it)
	db.Exec(`ALTER TABLE orders ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP`)
	db.Exec(`ALTER TABLE orders ADD COLUMN customer_city VARCHAR(255)`)
	db.Exec(`ALTER TABLE orders ADD COLUMN customer_state VARCHAR(100)`)
	db.Exec(`ALTER TABLE orders ADD COLUMN customer_country VARCHAR(100)`)

	_, err = db.Exec(`
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS subtotal_price DECIMAL(12, 2);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS total_tax DECIMAL(12, 2);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS store_id VARCHAR(255);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS external_order_id VARCHAR(255);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS source_id VARCHAR(50) DEFAULT 'shopify';
		ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_source_id_external_order_id_key;
		ALTER TABLE orders ADD CONSTRAINT orders_source_id_external_order_id_key UNIQUE(source_id, external_order_id);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS currency VARCHAR(10);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS financial_status VARCHAR(50);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS fulfillment_status VARCHAR(50);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP WITH TIME ZONE;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancel_reason TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS raw_payload JSONB;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_address1 TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_address2 TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_zip VARCHAR(20);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_first_name VARCHAR(255);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_last_name VARCHAR(255);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(50);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(100);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_company VARCHAR(100);
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_url TEXT;

		-- Add Foreign Key to Sources
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'orders_source_id_fkey') THEN
				ALTER TABLE orders ADD CONSTRAINT orders_source_id_fkey FOREIGN KEY (source_id) REFERENCES sources(id);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to update orders table schema: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS order_line_items (
			id VARCHAR(255) PRIMARY KEY,
			order_id VARCHAR(255) REFERENCES orders(id) ON DELETE CASCADE,
			title TEXT,
			sku VARCHAR(100),
			hs_code VARCHAR(50),
			quantity INTEGER,
			price DECIMAL(10, 2),
			discount DECIMAL(10, 2) DEFAULT 0,
			product_id VARCHAR(255),
			variant_id VARCHAR(255)
		);
		ALTER TABLE order_line_items ADD COLUMN IF NOT EXISTS discount DECIMAL(10, 2) DEFAULT 0;
		ALTER TABLE order_line_items ADD COLUMN IF NOT EXISTS product_id VARCHAR(255);
		ALTER TABLE order_line_items ADD COLUMN IF NOT EXISTS variant_id VARCHAR(255);
	`)

	// Create webhook_events table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS webhook_events (
			id SERIAL PRIMARY KEY,
			source_id VARCHAR(50) NOT NULL,
			order_id VARCHAR(255) REFERENCES orders(id) ON DELETE SET NULL,
			topic VARCHAR(100) NOT NULL,
			external_id VARCHAR(255) NOT NULL,
			webhook_delivery_id VARCHAR(255) UNIQUE NOT NULL,
			payload JSONB NOT NULL,
			processed BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		-- Add Foreign Key to Sources for Webhook Events
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'webhook_events_source_id_fkey') THEN
				ALTER TABLE webhook_events ADD CONSTRAINT webhook_events_source_id_fkey FOREIGN KEY (source_id) REFERENCES sources(id);
			END IF;
		END $$;
	`)
	// 7. Automation Engine Tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS automation_templates (
			id SERIAL PRIMARY KEY,
			store_id TEXT NOT NULL,
			template_name TEXT NOT NULL,
			language TEXT NOT NULL,
			category TEXT NOT NULL,
			body TEXT NOT NULL,
			header JSONB,
			footer TEXT,
			buttons JSONB,
			status TEXT DEFAULT 'pending',
			meta_template_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		ALTER TABLE automation_templates ADD COLUMN IF NOT EXISTS header JSONB;
		ALTER TABLE automation_templates ADD COLUMN IF NOT EXISTS footer TEXT;
		ALTER TABLE automation_templates ADD COLUMN IF NOT EXISTS buttons JSONB;

		CREATE TABLE IF NOT EXISTS automation_triggers (
			id SERIAL PRIMARY KEY,
			store_id TEXT NOT NULL,
			webhook_topic TEXT NOT NULL,
			template_id INTEGER REFERENCES automation_templates(id),
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS automation_messages (
			id SERIAL PRIMARY KEY,
			store_id TEXT NOT NULL,
			template_id INTEGER REFERENCES automation_templates(id),
			order_id TEXT NOT NULL,
			phone_number TEXT NOT NULL,
			message_id TEXT UNIQUE,
			status TEXT DEFAULT 'sent',
			sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			delivered_at TIMESTAMP,
			read_at TIMESTAMP,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS automation_whatsapp_settings (
			id SERIAL PRIMARY KEY,
			store_id VARCHAR(255) UNIQUE NOT NULL,
			enabled BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS webhook_status (
			id SERIAL PRIMARY KEY,
			topic VARCHAR(255),
			status VARCHAR(50),
			last_received TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		
		-- Insert a default row if not exists
		INSERT INTO webhook_status (id, topic, status, last_received)
		SELECT 1, 'none', 'inactive', NOW()
		WHERE NOT EXISTS (SELECT 1 FROM webhook_status WHERE id = 1);
	`)
	if err != nil {
		return fmt.Errorf("failed to create automation and setting tables: %w", err)
	}

	log.Println("Database auto-migration completed successfully")
	return nil
}

// TruncateOrders clears all orders from the local database
func TruncateOrders(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE TABLE order_line_items CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate order_line_items: %w", err)
	}
	_, err = db.Exec("TRUNCATE TABLE orders CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate orders: %w", err)
	}
	_, err = db.Exec("TRUNCATE TABLE webhook_events CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate webhook_events: %w", err)
	}
	log.Println("Successfully truncated orders, order_line_items, and webhook_events tables")
	return nil
}
