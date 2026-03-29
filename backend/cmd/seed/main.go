package main

import (
	"log"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"os"
)

func main() {
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	configsRepo := repository.NewConfigsRepository(db)
	settingsProvider := config.NewSettingsProvider(configsRepo)

	authService := service.NewAuthService(db, settingsProvider, nil)

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "password"
	}

	err = authService.Register(username, password)
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}

	log.Printf("Successfully registered user: %s with password: %s", username, password)
}
