package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"mi-tech/internal/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB initializes a test database connection.
// It uses a separate database name to avoid messing with local development data.
func SetupTestDB() (*gorm.DB, error) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	root := filepath.Join(basepath, "..", "..")

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/mi-tech-test?sslmode=disable"
	}

	// Create the test database if it doesn't exist
	if err := createTestDatabase(dsn); err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Run migrations from the migrations directory
	migrationsPath := filepath.Join(root, "internal/database/migrations")
	if err := runMigrations(db, migrationsPath); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func createTestDatabase(dsn string) error {
	parts := strings.Split(dsn, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid DSN format: %s", dsn)
	}

	lastPart := parts[len(parts)-1]
	dbNameParts := strings.Split(lastPart, "?")
	dbName := dbNameParts[0]

	// Create base DSN pointing to postgres
	baseParts := make([]string, len(parts)-1)
	copy(baseParts, parts[:len(parts)-1])
	baseDSN := strings.Join(baseParts, "/") + "/postgres"
	if len(dbNameParts) > 1 {
		baseDSN += "?" + dbNameParts[1]
	}

	db, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	var exists int
	db.Raw("SELECT 1 FROM pg_database WHERE datname = ?", dbName).Scan(&exists)
	if exists == 0 {
		if err := db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)).Error; err != nil {
			return err
		}
	}

	return nil
}

func runMigrations(db *gorm.DB, migrationsDir string) error {
	// 1. Change working directory temporarily to the migrations folder parent
	// because RunMigrations expects "internal/database/migrations" relative path
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	root := filepath.Join(basepath, "..", "..")

	originalWD, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(originalWD)

	if err := database.RunMigrations(db); err != nil {
		return err
	}
	return database.SeedDefaultUsers(db)
}

func CleanupTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
