package server

import (
	"context"
	"log"
	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/client/amazon"
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
	port      string
	mux       *http.ServeMux
	db        *gorm.DB
	amzPoller *service.AmazonOrderPoller
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
	inventoryRepo := repository.NewInventoryRepository(db)
	oilRepo := repository.NewOilInventoryRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	poRepo := repository.NewPurchaseOrderRepository(db)
	mfgRepo := repository.NewManufacturingRepository(db)
	aiReadRepo := repository.NewAIReadRepository(db)
	aiConvRepo := repository.NewAIConversationRepository(db)
	aiMemRepo := repository.NewAIMemoryRepository(db)

	// Providers
	settingsProvider := config.NewSettingsProvider(configsRepo)

	// Clients
	shopifyClient := shopify.NewClient(settingsProvider)
	amazonClient := amazon.NewClient(settingsProvider)

	// Orchestrators
	syncOrchestrator := service.NewSyncOrchestrator(inventoryRepo, shopifyClient, amazonClient, settingsProvider)

	// Services
	userService := service.NewUserService(db)
	metricsService := service.NewMetricsService(metricsRepo)
	reportService := service.NewReportService(reportRepo)
	customerService := service.NewCustomerService(customerRepo, orderRepo, shopifyClient)
	invoiceService := service.NewInvoiceService(settingsRepo)
	orderService := service.NewOrderService(orderRepo, lineItemRepo, customerService, shopifyClient, syncOrchestrator)
	syncService := service.NewSyncService(shopifyClient, orderRepo, customerService, syncOrchestrator)
	webhookService := service.NewWebhookService(orderService, shopifyClient, webhookEventRepo, webhookStatusRepo)
	amazonOrderPoller := service.NewAmazonOrderPoller(amazonClient, orderRepo, inventoryRepo, syncOrchestrator)
	plannerService := service.NewPlannerService(plannerRepo)
	inventoryService := service.NewInventoryService(inventoryRepo, shopifyClient, syncOrchestrator, settingsProvider, amazonOrderPoller)
	oilService := service.NewOilInventoryService(oilRepo)
	supplierService := service.NewSupplierService(supplierRepo)
	poService := service.NewPurchaseOrderService(poRepo, oilRepo)
	mfgService := service.NewManufacturingService(db, mfgRepo, oilRepo, syncOrchestrator)
	whatsappService := whatsapp.NewTemplatesService(whatsappRepo, settingsProvider)
	notifService := whatsapp.NewNotificationService(settingsProvider)
	agentService := whatsapp.NewAgentService(settingsProvider, plannerService, messagesRepo, whatsapp.NewMetaClient(settingsProvider), notifService)
	messagesService := whatsapp.NewMessagesService(messagesRepo, settingsProvider, customerRepo, agentService)
	authService := service.NewAuthService(db, settingsProvider, messagesService)
	metaMarketingClient := marketing.NewMetaMarketingClient(settingsProvider)
	socialService := service.NewSocialService(socialRepo, metaMarketingClient)
	systemService := service.NewSystemService("../docs")
	aiService := service.NewAIService(aiReadRepo, aiConvRepo, aiMemRepo, settingsProvider)
	marketingHandler := handler.NewMarketingHandler(metaMarketingClient)
	marketingWebhookHandler := handler.NewMarketingWebhookHandler(metaMarketingClient, settingsProvider)
	systemHandler := handler.NewSystemHandler(systemService)
	smmHandler := handler.NewSMMHandler(socialService)
	mappingService := whatsapp.NewWebhookMappingService(whatsappRepo, messagesService, invoiceService, settingsRepo, lineItemRepo, settingsProvider, orderRepo)

	// Handlers
	orderHandler := handler.NewOrderHandler(orderService, invoiceService, mappingService)
	syncHandler := handler.NewSyncHandler(syncService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	reportHandler := handler.NewReportHandler(reportService)
	webhookHandler := handler.NewWebhookHandler(webhookService, mappingService, settingsProvider)
	automationHandler := whatsapp.NewAutomationHandler(whatsappService, messagesService, mappingService, orderService, customerService, settingsProvider, agentService)
	settingsHandler := handler.NewSettingsHandler(settingsRepo)
	configsHandler := handler.NewConfigsHandler(configsRepo, db)
	redirectHandler := handler.NewRedirectHandler(orderRepo)
	authHandler := handler.NewAuthHandler(authService)
	customerHandler := handler.NewCustomerHandler(customerService)
	feedbackHandler := handler.NewFeedbackHandler(orderService, settingsProvider, mappingService, whatsappRepo)
	
	userHandler := handler.NewUserHandler(userService)
	plannerHandler := handler.NewPlannerHandler(plannerService, agentService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	oilHandler := handler.NewOilInventoryHandler(oilService)
	supplierHandler := handler.NewSupplierHandler(supplierService)
	poHandler := handler.NewPurchaseOrderHandler(poService)
	mfgHandler := handler.NewManufacturingHandler(mfgService)
	aiHandler := handler.NewAIHandler(aiService)

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
		feedbackHandler,
		inventoryHandler,
		oilHandler,
		supplierHandler,
		poHandler,
		mfgHandler,
		aiHandler,
		authService,
	)

	return &Server{
		port:      cfg.Port,
		mux:       mux,
		db:        db,
		amzPoller: amazonOrderPoller,
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

	// Start background workers
	if s.amzPoller != nil {
		go s.amzPoller.Start(context.Background())
	}

	return server.ListenAndServe()
}
