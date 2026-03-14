package main

import (
	"fmt"
	"log"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
)

func main() {
	cfg := config.Load()

	// Use ConnectDB instead of InitDB to avoid automatic migrations
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	fmt.Println("Database connection verified successfully (Ready-only check)")

	// Check if migrations table exists
	var exists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
	if !exists {
		fmt.Println("Warning: 'schema_migrations' table does not exist. The database may not have been initialized yet.")
		return
	}

	var count int64
	db.Table("schema_migrations").Count(&count)
	fmt.Printf("Total migrations tracked: %d\n", count)

	var files []string
	db.Table("schema_migrations").Select("filename").Order("filename ASC").Scan(&files)
	for _, f := range files {
		fmt.Printf("- %s\n", f)
	}
}
