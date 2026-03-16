package server

import (
	"log"
	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
	"mi-tech/internal/handler"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"net/http"

	"gorm.io/gorm"
)

type Server struct {
	port string
	mux  *http.ServeMux
	db   *gorm.DB
}

func New() (*Server, error) {
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		return nil, err
	}
	return NewServer(cfg, db), nil
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	mux := http.NewServeMux()

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Repositories
	orderRepo := repository.NewOrderRepository(db)
	lineItemRepo := repository.NewLineItemRepository(db)
	reportRepo := repository.NewReportRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	webhookEventRepo := repository.NewWebhookEventRepository(db)
	webhookStatusRepo := repository.NewWebhookStatusRepository(db)
	whatsappRepo := whatsapp.NewTemplatesRepository(sqlDB)
	messagesRepo := whatsapp.NewMessagesRepository(sqlDB)
	configsRepo := repository.NewConfigsRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Providers
	settingsProvider := config.NewSettingsProvider(configsRepo)

	// Clients
	shopifyClient := shopify.NewClient(settingsProvider)

	// Services
	invoiceService := service.NewInvoiceService(settingsRepo)
	orderService := service.NewOrderService(orderRepo, lineItemRepo)
	syncService := service.NewSyncService(shopifyClient, orderRepo)
	webhookService := service.NewWebhookService(orderService, shopifyClient, webhookEventRepo, webhookStatusRepo)
	whatsappService := whatsapp.NewTemplatesService(whatsappRepo, settingsProvider)
	messagesService := whatsapp.NewMessagesService(messagesRepo, settingsProvider)
	mappingService := whatsapp.NewWebhookMappingService(whatsappRepo, messagesService, invoiceService, settingsRepo)
	authService := service.NewAuthService(db, settingsProvider)
	metricsService := service.NewMetricsService(metricsRepo)
	reportService := service.NewReportService(reportRepo)

	// Handlers
	orderHandler := handler.NewOrderHandler(orderService, invoiceService)
	syncHandler := handler.NewSyncHandler(syncService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	reportHandler := handler.NewReportHandler(reportService)
	webhookHandler := handler.NewWebhookHandler(webhookService, mappingService, settingsProvider)
	automationHandler := whatsapp.NewAutomationHandler(whatsappService, messagesService, mappingService, orderService, settingsProvider)
	settingsHandler := handler.NewSettingsHandler(settingsRepo)
	configsHandler := handler.NewConfigsHandler(configsRepo, db)
	redirectHandler := handler.NewRedirectHandler(orderRepo)
	authHandler := handler.NewAuthHandler(authService)

	RegisterRoutes(
		mux,
		orderHandler,
		syncHandler,
		metricsHandler,
		reportHandler,
		webhookHandler,
		automationHandler,
		settingsHandler,
		configsHandler,
		redirectHandler,
		authHandler,
		authService,
	)

	return &Server{
		port: cfg.Port,
		mux:  mux,
		db:   db,
	}
}

func (s *Server) Run() error {
	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.mux,
	}

	log.Printf("Server starting on port %s", s.port)
	return server.ListenAndServe()
}
