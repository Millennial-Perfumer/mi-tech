package main

import (
	"fmt"
	"log"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
	"mi-tech/internal/entity"
)

func main() {
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var mapping entity.InventoryMapping
	sku := "QC-EJ8H-SMSV"
	if err := db.Where("platform = ? AND external_sku = ?", "amazon", sku).First(&mapping).Error; err != nil {
		fmt.Printf("Mapping for SKU %s NOT FOUND\n", sku)
	} else {
		fmt.Printf("SKU %s -> linked to Item ID %d\n", sku, mapping.InventoryItemID)
		var item entity.InventoryItem
		db.First(&item, mapping.InventoryItemID)
		fmt.Printf("Item Title: %s\n", item.Title)
	}
}
