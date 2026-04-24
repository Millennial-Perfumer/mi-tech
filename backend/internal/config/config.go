package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port  string
	DBDSN string

	// Amazon SP-API Config
	AmazonLWAClientID     string
	AmazonLWAClientSecret string
	AmazonLWARefreshToken string
	AmazonAWSAccessKey    string
	AmazonAWSSecretKey    string
	AmazonAWSRegion       string
	AmazonAWSRoleARN      string
	AmazonMarketplaceID   string
	AmazonSellerID        string
}

func (c *Config) GetDBDSN() string {
	return c.DBDSN
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

	// Prefer individual components for robustness against special characters
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	var dbDSN string
	if dbUser != "" && dbHost != "" {
		dbDSN = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			dbHost, dbUser, dbPass, dbName, dbPort)
	} else {
		// Fallback to DB_DSN if components are missing
		dbDSN = os.Getenv("DB_DSN")
	}

	return &Config{
		Port:  port,
		DBDSN: dbDSN,

		AmazonLWAClientID:     os.Getenv("AMAZON_LWA_CLIENT_ID"),
		AmazonLWAClientSecret: os.Getenv("AMAZON_LWA_CLIENT_SECRET"),
		AmazonLWARefreshToken: os.Getenv("AMAZON_LWA_REFRESH_TOKEN"),
		AmazonAWSAccessKey:    os.Getenv("AMAZON_AWS_ACCESS_KEY"),
		AmazonAWSSecretKey:    os.Getenv("AMAZON_AWS_SECRET_KEY"),
		AmazonAWSRegion:       os.Getenv("AMAZON_AWS_REGION"),
		AmazonAWSRoleARN:      os.Getenv("AMAZON_AWS_ROLE_ARN"),
		AmazonMarketplaceID:   os.Getenv("AMAZON_MARKETPLACE_ID"),
		AmazonSellerID:        os.Getenv("AMAZON_SELLER_ID"),
	}
}
