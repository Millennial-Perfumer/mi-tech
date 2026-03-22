package repository

import (
	"fmt"
	"os"
	"testing"
	"time"

	"mi-tech/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func BenchmarkUpsertLineItems(b *testing.B) {
	// Setup a temporary database for benchmarking if possible,
	// or use a mock. Since we want to measure DB performance,
	// a real DB is better.
	// For this environment, we might need to rely on the dockerized DB.
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=UTC"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Fatalf("Database not available for benchmark: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		b.Fatalf("Failed to get underlying *sql.DB: %v", err)
	}

	// Ensure the schema is ready for benchmarking
	if err := db.AutoMigrate(&entity.Order{}, &entity.LineItem{}); err != nil {
		b.Fatalf("Failed to auto-migrate database schema: %v", err)
	}

	// Ensure tables are cleaned up after the benchmark run
	b.Cleanup(func() {
		db.Exec("TRUNCATE TABLE order_line_items CASCADE")
		db.Exec("TRUNCATE TABLE orders CASCADE")
		sqlDB.Close()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Start a transaction for each iteration to isolate operations
		tx := db.Begin()
		if tx.Error != nil {
			b.Fatalf("Failed to begin transaction: %v", tx.Error)
		}
		// Use a repository initialized with the transaction
		txRepo := NewOrderRepository(tx)

		// Create a fresh order object for each iteration to ensure isolation
		currentOrder := entity.Order{
			SourceID:    "bench-source",
			OrderNumber: "BENCH001",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		currentOrder.ExternalOrderID = fmt.Sprintf("bench-order-%d", i) // Unique order ID

		// Create many line items
		for j := 0; j < 100; j++ {
			// Make LineItem IDs unique across all benchmark iterations
			id := fmt.Sprintf("li-%d-%d", i, j)
			title := fmt.Sprintf("Item %d", j)
			currentOrder.LineItems = append(currentOrder.LineItems, entity.LineItem{
				ID:       id,
				Title:    &title,
				Quantity: 1,
				Price:    10.0,
			})
		}

		err := txRepo.Upsert(currentOrder)
		if err != nil {
			tx.Rollback()
			b.Fatalf("Upsert failed: %v", err)
		}
		// Rollback the transaction to discard changes and ensure a consistent state for the next iteration
		tx.Rollback()
	}
}
