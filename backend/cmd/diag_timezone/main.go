package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/entity"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=password dbname=mi-tech port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var orders []entity.Order
	if err := db.Where("raw_payload IS NOT NULL").Order("created_at DESC").Limit(5).Find(&orders).Error; err != nil {
		log.Fatalf("failed to fetch orders: %v", err)
	}

	fmt.Println("ID | DB CreatedAt (UTC) | Payload CreatedAt | Diff (Hours)")
	fmt.Println("-----------------------------------------------------------")

	for _, o := range orders {
		var payload map[string]interface{}
		if err := json.Unmarshal(*o.RawPayload, &payload); err != nil {
			continue
		}

		payloadTimeStr, ok := payload["created_at"].(string)
		if !ok {
			continue
		}

		payloadTime, err := time.Parse(time.RFC3339, payloadTimeStr)
		if err != nil {
			continue
		}

		diff := o.CreatedAt.Sub(payloadTime.UTC()).Hours()
		fmt.Printf("%d | %s | %s | %.2f\n", o.ID, o.CreatedAt.Format(time.RFC3339), payloadTime.UTC().Format(time.RFC3339), diff)
	}
}
