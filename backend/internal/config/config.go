package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DBDSN                 string
	ShopifyStoreURL       string
	ShopifyAccessToken    string
	ShopifyWebhookSecret  string
	WhatsAppPhoneNumberID string
	WhatsAppAccessToken   string
	WhatsAppAppID         string
	WhatsAppAppSecret     string
	WhatsAppWABAID        string
	JWTSecret             string
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
		Port:                  port,
		DBDSN:                 dbDSN,
		ShopifyStoreURL:       os.Getenv("SHOPIFY_STORE_URL"),
		ShopifyAccessToken:    os.Getenv("SHOPIFY_ACCESS_TOKEN"),
		ShopifyWebhookSecret:  os.Getenv("SHOPIFY_WEBHOOK_SECRET"),
		WhatsAppPhoneNumberID: os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		WhatsAppAccessToken:   os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		WhatsAppAppID:         os.Getenv("WHATSAPP_APP_ID"),
		WhatsAppAppSecret:     os.Getenv("WHATSAPP_APP_SECRET"),
		WhatsAppWABAID:        os.Getenv("WHATSAPP_WABA_ID"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
	}
}
