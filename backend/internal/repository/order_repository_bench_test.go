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
		b.Skip("Database not available for benchmark")
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

	repo := NewOrderRepository(db)

	order := entity.Order{
		SourceID:    "bench-source",
		OrderNumber: "BENCH001",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create many line items
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("li-%d", i)
		title := fmt.Sprintf("Item %d", i)
		order.LineItems = append(order.LineItems, entity.LineItem{
			ID:       id,
			Title:    &title,
			Quantity: 1,
			Price:    10.0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Ensure each iteration is independent by providing a unique external order ID
		order.ExternalOrderID = fmt.Sprintf("bench-order-%d", i)
		err := repo.Upsert(order)
		if err != nil {
			b.Fatalf("Upsert failed: %v", err)
		}
	}
}
