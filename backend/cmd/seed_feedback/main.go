package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Order struct {
	ID               int64      `gorm:"primaryKey"`
	OrderNumber      string     `gorm:"column:order_number"`
	DeliveryStatus   string     `gorm:"column:delivery_status"`
	FeedbackStatusID int        `gorm:"column:feedback_status_id"`
	DeliveredAt      *time.Time `gorm:"column:delivered_at"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var order Order
	// Find the most recent delivered order
	err = db.Where("delivery_status = ?", "delivered").Order("id DESC").First(&order).Error
	if err != nil {
		log.Fatalf("No delivered orders found to update: %v", err)
	}

	// Backdate it to 6 days ago
	sixDaysAgo := time.Now().Add(-6 * 24 * time.Hour)
	err = db.Model(&order).Updates(map[string]interface{}{
		"delivered_at":       sixDaysAgo,
		"feedback_status_id": 1, // Reset to pending just in case
	}).Error

	if err != nil {
		log.Fatalf("Failed to update order: %v", err)
	}

	fmt.Printf("Successfully updated Order #%s (ID: %d) to be delivered at %v\n", 
		order.OrderNumber, order.ID, sixDaysAgo)
}
