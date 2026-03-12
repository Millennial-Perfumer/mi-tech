package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"shopify-gst-app/internal/config"

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

// runMigrations executes all SQL files in the migrations directory.
func runMigrations(db *gorm.DB) error {
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

		if err := db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	log.Println("All database migrations completed successfully")
	return nil
}
