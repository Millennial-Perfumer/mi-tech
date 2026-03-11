package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"shopify-gst-app/internal/config"

	_ "github.com/lib/pq"
)

// InitDB initializes the database connection pool and runs migrations.
func InitDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connection pool initialized")

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	return db, nil
}

// runMigrations executes all SQL files in the migrations directory.
func runMigrations(db *sql.DB) error {
	migrationsDir := "internal/database/migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("could not read migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		log.Printf("Executing migration: %s", filename)
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("could not read migration file %s: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	log.Println("All database migrations completed successfully")
	return nil
}

// TruncateOrders clears all orders from the local database.
// This is kept for backward compatibility with legacy sync code.
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
