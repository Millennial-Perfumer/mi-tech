package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/handler"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"

	"gorm.io/gorm"
)

// Server holds all dependencies and the HTTP server.
type Server struct {
	cfg      *config.Config
	database *gorm.DB
	httpSrv  *http.Server
}

// New initializes all dependencies and returns a ready-to-run Server.
func New() (*Server, error) {
	// 1. Config
	cfg := config.Load()

	// 2. Database
	db, err := database.InitDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 3. Repositories
	orderRepo := repository.NewOrderRepository(db)
	lineItemRepo := repository.NewLineItemRepository(db)
	webhookEventRepo := repository.NewWebhookEventRepository(db)
	webhookStatusRepo := repository.NewWebhookStatusRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	reportRepo := repository.NewReportRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// 4. External Clients
	shopifyClient := shopify.NewClient(cfg)

	// 5. Services
	orderService := service.NewOrderService(orderRepo, lineItemRepo)
	syncService := service.NewSyncService(shopifyClient, orderRepo)
	invoiceService := service.NewInvoiceService()
	metricsService := service.NewMetricsService(metricsRepo)
	reportService := service.NewReportService(reportRepo)
	webhookService := service.NewWebhookService(orderService, shopifyClient, webhookEventRepo, webhookStatusRepo)
	authService := service.NewAuthService(db, cfg.JWTSecret)

	// 6. WhatsApp Automation Module
	sqlDB, _ := db.DB()
	templatesRepo := whatsapp.NewTemplatesRepository(sqlDB)
	messagesRepo := whatsapp.NewMessagesRepository(sqlDB)
	templatesService := whatsapp.NewTemplatesService(templatesRepo, cfg)
	messagesService := whatsapp.NewMessagesService(messagesRepo, cfg)
	mappingService := whatsapp.NewWebhookMappingService(templatesRepo, messagesService, invoiceService)
	automationHandler := whatsapp.NewAutomationHandler(templatesService, messagesService)

	// 7. Handlers
	orderHandler := handler.NewOrderHandler(orderService, invoiceService)
	syncHandler := handler.NewSyncHandler(syncService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	reportHandler := handler.NewReportHandler(reportService)
	webhookHandler := handler.NewWebhookHandler(webhookService, mappingService, cfg.ShopifyWebhookSecret)
	settingsHandler := handler.NewSettingsHandler(settingsRepo)
	redirectHandler := handler.NewRedirectHandler(orderRepo)
	authHandler := handler.NewAuthHandler(authService)

	// 8. Router
	mux := http.NewServeMux()
	RegisterRoutes(mux, orderHandler, syncHandler, metricsHandler, reportHandler, webhookHandler, automationHandler, settingsHandler, redirectHandler, authHandler, authService)

	// 9. HTTP Server with timeouts
	httpSrv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:      cfg,
		database: db,
		httpSrv:  httpSrv,
	}, nil
}

// Run starts the HTTP server and handles graceful shutdown.
func (s *Server) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting backend server on %s...", s.httpSrv.Addr)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Received signal interrupt. Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	sqlDB, _ := s.database.DB()
	sqlDB.Close()
	log.Println("Server stopped cleanly.")
	return nil
}
