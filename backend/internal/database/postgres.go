package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mi-tech/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the GORM database connection and runs migrations.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connection pool initialized")

	// Run migrations (keep SQL-file-based migrations for schema)
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	return db, nil
}

// runMigrations executes all SQL files in the migrations directory with tracking.
func runMigrations(db *gorm.DB) error {
	migrationsDir := "internal/database/migrations"

	// 1. Ensure migrations table exists
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`).Error
	if err != nil {
		return fmt.Errorf("could not create migrations table: %w", err)
	}

	// 2. Load already applied migrations
	var applied []string
	if err := db.Raw("SELECT filename FROM schema_migrations").Scan(&applied).Error; err != nil {
		return fmt.Errorf("could not load applied migrations: %w", err)
	}
	appliedMap := make(map[string]bool)
	for _, f := range applied {
		appliedMap[f] = true
	}

	// 3. Read migration files
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

	// 4. Execute new migrations in transactions
	for _, filename := range sqlFiles {
		if appliedMap[filename] {
			continue
		}

		log.Printf("Applying migration: %s", filename)
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("could not read migration file %s: %w", filename, err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(content)).Error; err != nil {
				return err
			}
			return tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", filename).Error
		})

		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	log.Println("Database migration check completed")
	return nil
}
