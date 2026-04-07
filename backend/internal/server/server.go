package server

import (
	"log"
	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/config"
	"mi-tech/internal/database"
	"mi-tech/internal/handler"
	"mi-tech/internal/marketing"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	customerRepo := repository.NewCustomerRepository(db)
	socialRepo := repository.NewSocialRepository(db)
	plannerRepo := repository.NewPlannerRepository(db)

	// Providers
	settingsProvider := config.NewSettingsProvider(configsRepo)

	// Clients
	shopifyClient := shopify.NewClient(settingsProvider)

	// Services
	userService := service.NewUserService(db)
	metricsService := service.NewMetricsService(metricsRepo)
	reportService := service.NewReportService(reportRepo)
	customerService := service.NewCustomerService(customerRepo, orderRepo, shopifyClient)
	invoiceService := service.NewInvoiceService(settingsRepo)
	orderService := service.NewOrderService(orderRepo, lineItemRepo, customerService, shopifyClient)
	syncService := service.NewSyncService(shopifyClient, orderRepo, customerService)
	webhookService := service.NewWebhookService(orderService, shopifyClient, webhookEventRepo, webhookStatusRepo)
	whatsappService := whatsapp.NewTemplatesService(whatsappRepo, settingsProvider)
	messagesService := whatsapp.NewMessagesService(messagesRepo, settingsProvider, customerRepo)
	authService := service.NewAuthService(db, settingsProvider, messagesService)
	metaMarketingClient := marketing.NewMetaMarketingClient(settingsProvider)
	socialService := service.NewSocialService(socialRepo, metaMarketingClient)
	plannerService := service.NewPlannerService(plannerRepo)
	systemService := service.NewSystemService("../docs")
	marketingHandler := handler.NewMarketingHandler(metaMarketingClient)
	marketingWebhookHandler := handler.NewMarketingWebhookHandler(metaMarketingClient, settingsProvider)
	systemHandler := handler.NewSystemHandler(systemService)
	smmHandler := handler.NewSMMHandler(socialService)
	mappingService := whatsapp.NewWebhookMappingService(whatsappRepo, messagesService, invoiceService, settingsRepo, lineItemRepo)

	// Handlers
	orderHandler := handler.NewOrderHandler(orderService, invoiceService)
	syncHandler := handler.NewSyncHandler(syncService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	reportHandler := handler.NewReportHandler(reportService)
	webhookHandler := handler.NewWebhookHandler(webhookService, mappingService, settingsProvider)
	automationHandler := whatsapp.NewAutomationHandler(whatsappService, messagesService, mappingService, orderService, customerService, settingsProvider)
	settingsHandler := handler.NewSettingsHandler(settingsRepo)
	configsHandler := handler.NewConfigsHandler(configsRepo, db)
	redirectHandler := handler.NewRedirectHandler(orderRepo)
	authHandler := handler.NewAuthHandler(authService)
	customerHandler := handler.NewCustomerHandler(customerService)
	userHandler := handler.NewUserHandler(userService)
	plannerHandler := handler.NewPlannerHandler(plannerService)

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
		customerHandler,
		userHandler,
		marketingHandler,
		marketingWebhookHandler,
		systemHandler,
		smmHandler,
		plannerHandler,
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
		Addr:              ":" + s.port,
		Handler:           otelhttp.NewHandler(s.mux, "mi-tech-api"),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server starting on port %s", s.port)
	return server.ListenAndServe()
}
