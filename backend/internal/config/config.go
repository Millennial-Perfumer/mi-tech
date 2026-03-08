package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DBDSN              string
	ShopifyStoreURL    string
	ShopifyAccessToken string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, falling back to environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://postgres:password@localhost:5432/wisegst?sslmode=disable"
	}

	return &Config{
		Port:               port,
		DBDSN:              dbDSN,
		ShopifyStoreURL:    os.Getenv("SHOPIFY_STORE_URL"),
		ShopifyAccessToken: os.Getenv("SHOPIFY_ACCESS_TOKEN"),
	}
}
