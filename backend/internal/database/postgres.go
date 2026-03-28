package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig is a subset of config.Config to avoid circular dependencies
type DBConfig interface {
	GetDBDSN() string
}

// InitDB initializes the database and runs migrations.
func InitDB(cfg DBConfig) (*gorm.DB, error) {
	db, err := ConnectDB(cfg)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	if err := SeedDefaultUsers(db); err != nil {
		return nil, fmt.Errorf("user seeding failure: %w", err)
	}

	return db, nil
}

// SeedDefaultUsers creates the initial admin and read users if they don't exist
func SeedDefaultUsers(db *gorm.DB) error {
	users := []struct {
		Email string
		Role  string
		Pass  string
	}{
		{"admin@millennialperfumer.in", "admin", "admin123"},
		{"mi-agents@millennialperfumer.in", "read", "agent123"},
	}

	for _, u := range users {
		var hash string
		if u.Pass == "admin123" {
			hash = "$2a$10$TvklUXYlW7ysyiH9tIxJ6uejoEeftduQo7.sOue9dLbtyefSkmpZK" // admin123
		} else {
			hash = "$2a$10$4N5zIhXY3IG8h8Ax/z6fluoKbqoupMBdjstIDXNjeL6XZOJ6xs7Ru" // agent123
		}

		err := db.Exec(`
			INSERT INTO users (username, password_hash, role, created_at, updated_at) 
			VALUES (?, ?, ?, NOW(), NOW())
			ON CONFLICT (username) DO UPDATE 
			SET role = EXCLUDED.role, password_hash = EXCLUDED.password_hash, updated_at = NOW()
		`, u.Email, hash, u.Role).Error

		if err != nil {
			return err
		}
		log.Printf("Seeded and updated %s user: %s", u.Role, u.Email)
	}
	return nil
}

// ConnectDB initializes the GORM database connection without running migrations.
func ConnectDB(cfg DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDBDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Force session to UTC to avoid double-offset issues with IST
	if err := db.Exec("SET TimeZone='UTC'").Error; err != nil {
		log.Printf("Warning: Failed to set database timezone to UTC: %v", err)
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
	return db, nil
}

// RunMigrations executes all SQL files in the migrations directory with tracking.
func RunMigrations(db *gorm.DB) error {
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

	// --- Bootstrapping Check ---
	// If schema_migrations is empty, check if core tables already exist.
	// This prevents re-running migrations 001-008 in existing environments.
	if len(applied) == 0 {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'orders')").Scan(&exists)
		if exists {
			log.Println("Existing installation detected. Bootstrapping migration tracking...")
			files, err := os.ReadDir(migrationsDir)
			if err != nil {
				log.Printf("Warning: Failed to read migrations directory for bootstrapping: %v", err)
			} else {
				for _, f := range files {
					name := f.Name()
					// Only bootstrap 001-008 (the ones that existed before tracking)
					if strings.HasSuffix(name, ".sql") && name < "009" {
						db.Exec("INSERT INTO schema_migrations (filename) VALUES (?) ON CONFLICT DO NOTHING", name)
						applied = append(applied, name)
					}
				}
			}
		}
	}

	appliedMap := make(map[string]bool)
	for _, f := range applied {
		appliedMap[f] = true
	}

	// 3. Read migration files
	files, err := os.ReadDir(migrationsDir)
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
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
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
