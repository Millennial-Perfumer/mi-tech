package main

import (
	"fmt"
	"log"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
)

func main() {
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
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
