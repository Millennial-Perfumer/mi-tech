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
	query := `
	CREATE TABLE IF NOT EXISTS shopify_orders (
		id VARCHAR(255) PRIMARY KEY,
		order_number VARCHAR(255) NOT NULL,
		total_price DECIMAL(12, 2) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		customer_name VARCHAR(255),
		customer_city VARCHAR(255),
		customer_state VARCHAR(100),
		customer_country VARCHAR(100),
		status VARCHAR(100)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Gracefully add columns to existing tables (will error innocuously if already exists, so we ignore it)
	db.Exec(`ALTER TABLE shopify_orders ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP`)
	db.Exec(`ALTER TABLE shopify_orders ADD COLUMN customer_city VARCHAR(255)`)
	db.Exec(`ALTER TABLE shopify_orders ADD COLUMN customer_state VARCHAR(100)`)
	db.Exec(`ALTER TABLE shopify_orders ADD COLUMN customer_country VARCHAR(100)`)

	_, err = db.Exec(`
		ALTER TABLE shopify_orders ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255);
		ALTER TABLE shopify_orders ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50);
		ALTER TABLE shopify_orders ADD COLUMN IF NOT EXISTS subtotal_price DECIMAL(12, 2);
		ALTER TABLE shopify_orders ADD COLUMN IF NOT EXISTS total_tax DECIMAL(12, 2);
	`)
	if err != nil {
		return fmt.Errorf("failed to update shopify_orders table schema: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS shopify_order_line_items (
			id VARCHAR(255) PRIMARY KEY,
			order_id VARCHAR(255) REFERENCES shopify_orders(id) ON DELETE CASCADE,
			title TEXT,
			sku VARCHAR(100),
			hs_code VARCHAR(50),
			quantity INTEGER,
			price DECIMAL(10, 2),
			discount DECIMAL(10, 2) DEFAULT 0
		);
		ALTER TABLE shopify_order_line_items ADD COLUMN IF NOT EXISTS discount DECIMAL(10, 2) DEFAULT 0;
	`)
	if err != nil {
		return fmt.Errorf("failed to create shopify_order_line_items table: %w", err)
	}

	log.Println("Database auto-migration completed successfully")
	return nil
}

// TruncateOrders clears all orders from the local database
func TruncateOrders(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE TABLE shopify_order_line_items CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate shopify_order_line_items: %w", err)
	}
	_, err = db.Exec("TRUNCATE TABLE shopify_orders CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate shopify_orders: %w", err)
	}
	log.Println("Successfully truncated shopify_orders and shopify_order_line_items tables")
	return nil
}
