package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found, using system env")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("FAILED to connect to DB: %v", err)
	}

	var config struct {
		Key   string `gorm:"column:key"`
		Value string `gorm:"column:value"`
	}
	if err := db.Table("app_configs").Where("key = ?", "meta_marketing_access_token").First(&config).Error; err != nil {
		log.Fatalf("FAILED to get token: %v", err)
	}

	token := config.Value
	fmt.Printf("RESULT: Current Token in DB starts with: %s...\n", token[:20])

	// Fetch Pages
	u := fmt.Sprintf("https://graph.facebook.com/v22.0/me/accounts?access_token=%s&fields=id,name", token)
	resp, err := http.Get(u)
	if err != nil {
		log.Fatalf("FAILED to fetch pages: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("FAILED to decode pages: %v", err)
	}

	if result.Error != nil {
		fmt.Printf("META ERROR: %s\n", result.Error.Message)
		return
	}

	fmt.Println("\n--- AVAILABLE PAGES FOR THIS TOKEN ---")
	for _, p := range result.Data {
		fmt.Printf("ID: %s | NAME: %s\n", p.ID, p.Name)
	}
}
