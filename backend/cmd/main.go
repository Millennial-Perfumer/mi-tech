package main

import (
	"fmt"
	"log"
	"net/http"

	"shopify-gst-app/internal/automation/whatsapp"
	"shopify-gst-app/internal/config"
	"shopify-gst-app/internal/db"
	"shopify-gst-app/internal/handlers"
	"shopify-gst-app/internal/orders"
	"shopify-gst-app/internal/shopify"
)

// corsMiddleware adds basic CORS headers to requests
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	port := ":" + cfg.Port

	// 1. Initialize Orders Module
	ordersRepo := orders.NewRepository(database)
	ordersService := orders.NewService(ordersRepo)

	// 2. Initialize WhatsApp Automation Module
	templatesRepo := whatsapp.NewTemplatesRepository(database)
	messagesRepo := whatsapp.NewMessagesRepository(database)

	templatesService := whatsapp.NewTemplatesService(templatesRepo, cfg)
	messagesService := whatsapp.NewMessagesService(messagesRepo, cfg)
	mappingService := whatsapp.NewWebhookMappingService(templatesRepo, messagesService)
	automationHandler := whatsapp.NewAutomationHandler(templatesService, messagesService)

	// 3. Initialize Shopify Services & Handlers
	shopifyClient := shopify.NewClient(cfg)
	syncService := shopify.NewSyncService(shopifyClient, database)
	ordersHandler := handlers.NewOrdersHandler(database, syncService)
	metricsHandler := handlers.NewMetricsHandler(database)
	reportsHandler := handlers.NewReportsHandler(database)

	// 4. Initialize Webhook Handler
	webhooksHandler := handlers.NewWebhooksHandler(ordersService, mappingService, cfg.ShopifyWebhookSecret, database)

	// Register API Routes with CORS
	http.Handle("/api/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "GST Invoice Manager API is running"}`))
	}))

	http.HandleFunc("/api/dashboard/metrics", corsMiddleware(metricsHandler.GetDashboardMetrics))
	http.HandleFunc("/api/orders", corsMiddleware(ordersHandler.GetOrders))
	http.HandleFunc("/api/orders/status", corsMiddleware(ordersHandler.UpdateOrderStatus))
	http.HandleFunc("/api/orders/invoice", corsMiddleware(ordersHandler.GenerateInvoice))
	http.HandleFunc("/api/shopify/sync", corsMiddleware(ordersHandler.SyncOrders))
	http.HandleFunc("/api/shopify/reset", corsMiddleware(ordersHandler.ResetOrders))
	http.HandleFunc("/api/reports/summary", corsMiddleware(reportsHandler.GetGSTSummary))
	http.HandleFunc("/api/reports/state-wise", corsMiddleware(reportsHandler.GetStateSummary))
	http.HandleFunc("/api/reports/hsn-wise", corsMiddleware(reportsHandler.GetHSNSummary))
	http.HandleFunc("/api/reports/documents-issued", corsMiddleware(reportsHandler.GetDocumentsIssued))

	// Webhook Status Route
	http.HandleFunc("/api/webhook/status", corsMiddleware(webhooksHandler.GetWebhookStatus))

	// WhatsApp Automation Engine Routes
	http.HandleFunc("/api/automation/whatsapp/metrics", corsMiddleware(automationHandler.GetAutomationMetrics))
	http.HandleFunc("/api/automation/whatsapp/templates", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			automationHandler.CreateTemplate(w, r)
		case http.MethodPut:
			automationHandler.UpdateTemplate(w, r)
		case http.MethodDelete:
			automationHandler.DeleteTemplate(w, r)
		default:
			automationHandler.GetTemplates(w, r)
		}
	}))
	http.HandleFunc("/api/automation/whatsapp/triggers", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			automationHandler.CreateTrigger(w, r)
		case http.MethodPut:
			automationHandler.UpdateTrigger(w, r)
		case http.MethodDelete:
			automationHandler.DeleteTrigger(w, r)
		default:
			automationHandler.GetTriggers(w, r)
		}
	}))
	http.HandleFunc("/api/automation/whatsapp/messages", corsMiddleware(automationHandler.GetMessages))
	http.HandleFunc("/api/automation/whatsapp/webhook", automationHandler.WhatsAppWebhook)
	http.HandleFunc("/api/automation/whatsapp/debug/templates", corsMiddleware(automationHandler.DebugGetTemplates))

	// Main Webhook Route
	http.HandleFunc("/webhooks/shopify", webhooksHandler.ShopifyWebhookHandler)

	fmt.Printf("Starting backend server on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
