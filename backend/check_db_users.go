package main

import (
	"fmt"
	"log"
	"mi-tech/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=password dbname=mi-tech port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var users []entity.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("failed to query users: %v", err)
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, u := range users {
		fmt.Printf("- ID: %d, Username: %s, Role: %s\n", u.ID, u.Username, u.Role)
	}
}
