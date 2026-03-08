package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// ShopifyWebhookTestHandler handles incoming webhooks from Shopify for testing PII fields.
func ShopifyWebhookTestHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook Error: Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	log.Println("--- NEW SHOPIFY WEBHOOK PAYLOAD ---")
	log.Println(string(body))

	// Attempt to parse some key fields for quick logging
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		log.Printf("Order ID: %v", payload["id"])
		log.Printf("Order Number: %v", payload["order_number"])
		log.Printf("Email: %v", payload["email"])

		if cust, ok := payload["customer"].(map[string]interface{}); ok {
			log.Printf("Customer First Name: %v", cust["first_name"])
			log.Printf("Customer Last Name: %v", cust["last_name"])
		}

		if ship, ok := payload["shipping_address"].(map[string]interface{}); ok {
			log.Printf("Shipping Address Name: %v", ship["name"])
			log.Printf("Shipping Address 1: %v", ship["address1"])
			log.Printf("Shipping Address City: %v", ship["city"])
			log.Printf("Shipping Address Province: %v", ship["province"])
			log.Printf("Shipping Address Country: %v", ship["country"])
		}
	} else {
		log.Printf("Webhook Warning: Failed to parse JSON for field extraction: %v", err)
	}

	log.Println("--- END WEBHOOK PAYLOAD ---")

	w.WriteHeader(http.StatusOK)
}
