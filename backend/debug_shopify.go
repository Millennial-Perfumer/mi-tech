package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"mi-tech/internal/client/shopify"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
)

func main() {
	_, db, err := database.NewDatabase("postgres://postgres:password@localhost:5432/mi-tech?sslmode=disable")
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	settingsProvider := config.NewDatabaseSettingsProvider(db)
	client := shopify.NewClient(settingsProvider)

	order, err := client.FetchOrderByID("6877117645090")
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}

	b, _ := json.MarshalIndent(order.Fulfillments, "", "  ")
	fmt.Println(string(b))
}
