package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shopify-gst-app/internal/automation/whatsapp"
	"shopify-gst-app/internal/config"
	"shopify-gst-app/internal/database"
	"shopify-gst-app/internal/handler"
	"shopify-gst-app/internal/repository"
	"shopify-gst-app/internal/service"
	"shopify-gst-app/internal/client/shopify"
)

// Server holds all dependencies and the HTTP server.
type Server struct {
	cfg      *config.Config
	database *sql.DB
	httpSrv  *http.Server
}

// New initializes all dependencies and returns a ready-to-run Server.
func New() (*Server, error) {
	// 1. Config
	cfg := config.Load()

	// 2. Database
	database, err := database.InitDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 3. Repositories
	orderRepo := repository.NewOrderRepository(database)
	lineItemRepo := repository.NewLineItemRepository(database)
	webhookEventRepo := repository.NewWebhookEventRepository(database)
	webhookStatusRepo := repository.NewWebhookStatusRepository(database)
	metricsRepo := repository.NewMetricsRepository(database)
	reportRepo := repository.NewReportRepository(database)

	// 4. External Clients
	shopifyClient := shopify.NewClient(cfg)

	// 5. Services
	orderService := service.NewOrderService(orderRepo, lineItemRepo)
	syncService := service.NewSyncService(shopifyClient, orderRepo)
	invoiceService := service.NewInvoiceService()
	metricsService := service.NewMetricsService(metricsRepo)
	reportService := service.NewReportService(reportRepo)
	webhookService := service.NewWebhookService(orderService, webhookEventRepo, webhookStatusRepo)

	// 6. WhatsApp Automation Module (uses old patterns until Phase 7)
	templatesRepo := whatsapp.NewTemplatesRepository(database)
	messagesRepo := whatsapp.NewMessagesRepository(database)
	templatesService := whatsapp.NewTemplatesService(templatesRepo, cfg)
	messagesService := whatsapp.NewMessagesService(messagesRepo, cfg)
	mappingService := whatsapp.NewWebhookMappingService(templatesRepo, messagesService)
	automationHandler := whatsapp.NewAutomationHandler(templatesService, messagesService)

	// 7. Handlers
	orderHandler := handler.NewOrderHandler(orderService, invoiceService)
	syncHandler := handler.NewSyncHandler(syncService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	reportHandler := handler.NewReportHandler(reportService)
	webhookHandler := handler.NewWebhookHandler(webhookService, mappingService, cfg.ShopifyWebhookSecret)

	// 8. Router
	mux := http.NewServeMux()
	RegisterRoutes(mux, orderHandler, syncHandler, metricsHandler, reportHandler, webhookHandler, automationHandler)

	// 9. HTTP Server with timeouts
	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:      cfg,
		database: database,
		httpSrv:  httpSrv,
	}, nil
}

// Run starts the HTTP server and handles graceful shutdown on OS signals.
func (s *Server) Run() error {
	defer s.database.Close()

	// Channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting backend server on %s...", s.httpSrv.Addr)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Block until we receive a signal
	sig := <-quit
	log.Printf("Received signal %s. Shutting down gracefully...", sig)

	// Give active connections 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server stopped cleanly.")
	return nil
}
