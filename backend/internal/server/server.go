package server

import (
	"context"
	"log"
	aiHandlerPkg "mi-tech/internal/domain/ai/handler"
	aiRepoPkg "mi-tech/internal/domain/ai/repository"
	aiServicePkg "mi-tech/internal/domain/ai/service"
	communicationHandlerPkg "mi-tech/internal/domain/communication/handler"
	communicationRepoPkg "mi-tech/internal/domain/communication/repository"
	communicationServicePkg "mi-tech/internal/domain/communication/service"
	dashboardHandlerPkg "mi-tech/internal/domain/dashboard/handler"
	dashboardRepoPkg "mi-tech/internal/domain/dashboard/repository"
	dashboardServicePkg "mi-tech/internal/domain/dashboard/service"
	feedbackHandlerPkg "mi-tech/internal/domain/feedback/handler"
	feedbackRepoPkg "mi-tech/internal/domain/feedback/repository"
	feedbackServicePkg "mi-tech/internal/domain/feedback/service"
	gstHandlerPkg "mi-tech/internal/domain/gst/handler"
	gstRepoPkg "mi-tech/internal/domain/gst/repository"
	gstServicePkg "mi-tech/internal/domain/gst/service"
	inventoryHandlerPkg "mi-tech/internal/domain/inventory/handler"
	inventoryRepoPkg "mi-tech/internal/domain/inventory/repository"
	inventoryServicePkg "mi-tech/internal/domain/inventory/service"
	marketingHandlerPkg "mi-tech/internal/domain/marketing/handler"
	marketingRepoPkg "mi-tech/internal/domain/marketing/repository"
	marketingServicePkg "mi-tech/internal/domain/marketing/service"
	orderHandlerPkg "mi-tech/internal/domain/order/handler"
	orderRepoPkg "mi-tech/internal/domain/order/repository"
	orderServicePkg "mi-tech/internal/domain/order/service"
	plannerHandlerPkg "mi-tech/internal/domain/planner/handler"
	plannerRepoPkg "mi-tech/internal/domain/planner/repository"
	plannerServicePkg "mi-tech/internal/domain/planner/service"
	productionHandlerPkg "mi-tech/internal/domain/production/handler"
	productionRepoPkg "mi-tech/internal/domain/production/repository"
	productionServicePkg "mi-tech/internal/domain/production/service"
	supportHandlerPkg "mi-tech/internal/domain/support/handler"
	supportRepoPkg "mi-tech/internal/domain/support/repository"
	supportServicePkg "mi-tech/internal/domain/support/service"
	syncHandlerPkg "mi-tech/internal/domain/sync/handler"
	syncServicePkg "mi-tech/internal/domain/sync/service"
	userHandlerPkg "mi-tech/internal/domain/user/handler"
	userRepoPkg "mi-tech/internal/domain/user/repository"
	userServicePkg "mi-tech/internal/domain/user/service"
	webhookHandlerPkg "mi-tech/internal/domain/webhook/handler"
	webhookRepoPkg "mi-tech/internal/domain/webhook/repository"
	webhookServicePkg "mi-tech/internal/domain/webhook/service"
	b2bHandlerPkg "mi-tech/internal/domain/b2b/handler"
	b2bRepoPkg "mi-tech/internal/domain/b2b/repository"
	b2bServicePkg "mi-tech/internal/domain/b2b/service"
	"mi-tech/internal/shared/config"
	configHandlerPkg "mi-tech/internal/shared/config/handler"
	configRepoPkg "mi-tech/internal/shared/config/repository"
	"mi-tech/internal/shared/database"
	"mi-tech/internal/shared/extclient/amazon"
	"mi-tech/internal/shared/extclient/shopify"
	systemHandlerPkg "mi-tech/internal/shared/system/handler"
	systemServicePkg "mi-tech/internal/shared/system/service"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"gorm.io/gorm"
)

type Server struct {
	port              string
	mux               *http.ServeMux
	db                *gorm.DB
	amzPoller         *syncServicePkg.AmazonOrderPoller
	feedbackScheduler *feedbackServicePkg.FeedbackScheduler
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
	orderRepo := orderRepoPkg.NewOrderRepository(db)
	lineItemRepo := orderRepoPkg.NewLineItemRepository(db)
	reportRepo := gstRepoPkg.NewGSTRepository(db)
	metricsRepo := dashboardRepoPkg.NewMetricsRepository(db)
	webhookEventRepo := webhookRepoPkg.NewWebhookEventRepository(db)
	webhookStatusRepo := webhookRepoPkg.NewWebhookStatusRepository(db)
	whatsappRepo := communicationRepoPkg.NewTemplatesRepository(sqlDB)
	messagesRepo := communicationRepoPkg.NewMessagesRepository(sqlDB)
	configsRepo := configRepoPkg.NewConfigsRepository(db)
	settingsRepo := configRepoPkg.NewSettingsRepository(db)
	customerRepo := orderRepoPkg.NewCustomerRepository(db)
	socialRepo := marketingRepoPkg.NewSocialRepository(db)
	plannerRepo := plannerRepoPkg.NewPlannerRepository(db)
	inventoryRepo := inventoryRepoPkg.NewInventoryRepository(db)
	oilRepo := productionRepoPkg.NewOilInventoryRepository(db)
	supplierRepo := productionRepoPkg.NewSupplierRepository(db)
	poRepo := productionRepoPkg.NewPurchaseOrderRepository(db)
	mfgRepo := productionRepoPkg.NewManufacturingRepository(db)
	feedbackRepo := feedbackRepoPkg.NewFeedbackRepository(db)
	userRepo := userRepoPkg.NewUserRepository(db)
	ticketRepo := supportRepoPkg.NewTicketRepository(db)
	aiReadRepo := aiRepoPkg.NewAIReadRepository(db)
	aiConvRepo := aiRepoPkg.NewAIConversationRepository(db)
	aiMemRepo := aiRepoPkg.NewAIMemoryRepository(db)
	b2bRepo := b2bRepoPkg.NewB2BRepository(db)

	// Providers
	settingsProvider := config.NewSettingsProvider(configsRepo)

	// Clients
	shopifyClient := shopify.NewClient(settingsProvider)
	amazonClient := amazon.NewClient(settingsProvider)

	// Orchestrators
	syncOrchestrator := syncServicePkg.NewSyncOrchestrator(inventoryRepo, shopifyClient, amazonClient, settingsProvider)

	// Services
	userService := userServicePkg.NewUserService(userRepo)
	metricsService := dashboardServicePkg.NewMetricsService(metricsRepo)
	reportService := gstServicePkg.NewGSTService(reportRepo)
	customerService := orderServicePkg.NewCustomerService(customerRepo, orderRepo, shopifyClient)
	invoiceService := orderServicePkg.NewInvoiceService(settingsRepo)
	orderService := orderServicePkg.NewOrderService(orderRepo, lineItemRepo, customerService, shopifyClient, syncOrchestrator)
	feedbackService := feedbackServicePkg.NewFeedbackService(feedbackRepo, orderService)
	syncService := syncServicePkg.NewSyncService(shopifyClient, orderRepo, customerService, syncOrchestrator)
	webhookService := webhookServicePkg.NewWebhookService(orderService, shopifyClient, webhookEventRepo, webhookStatusRepo)
	amazonOrderPoller := syncServicePkg.NewAmazonOrderPoller(amazonClient, orderRepo, inventoryRepo, syncOrchestrator)
	plannerService := plannerServicePkg.NewPlannerService(plannerRepo)
	ticketService := supportServicePkg.NewTicketService(ticketRepo)
	inventoryService := inventoryServicePkg.NewInventoryService(inventoryRepo, shopifyClient, syncOrchestrator, settingsProvider, amazonOrderPoller)
	oilService := productionServicePkg.NewOilInventoryService(oilRepo)
	supplierService := productionServicePkg.NewSupplierService(supplierRepo)
	poService := productionServicePkg.NewPurchaseOrderService(poRepo, oilRepo)
	mfgService := productionServicePkg.NewManufacturingService(db, mfgRepo, oilRepo, syncOrchestrator)
	whatsappService := communicationServicePkg.NewTemplatesService(whatsappRepo, settingsProvider)
	notifService := communicationServicePkg.NewNotificationService(settingsProvider)
	agentService := communicationServicePkg.NewNewAgentService(settingsProvider, plannerService, ticketService, messagesRepo, communicationServicePkg.NewMetaClient(settingsProvider), notifService)
	messagesService := communicationServicePkg.NewMessagesService(messagesRepo, settingsProvider, customerRepo, agentService)
	authService := userServicePkg.NewAuthService(userRepo, settingsProvider, messagesService)
	metaMarketingClient := marketingServicePkg.NewMetaMarketingClient(settingsProvider)
	socialService := marketingServicePkg.NewSocialService(socialRepo, metaMarketingClient)
	systemService := systemServicePkg.NewSystemService("../docs")
	aiService := aiServicePkg.NewAIService(aiReadRepo, aiConvRepo, aiMemRepo, settingsProvider)
	b2bService := b2bServicePkg.NewB2BService(b2bRepo, settingsProvider, db)
	marketingHandler := marketingHandlerPkg.NewMarketingHandler(metaMarketingClient)
	marketingWebhookHandler := marketingHandlerPkg.NewMarketingWebhookHandler(metaMarketingClient, settingsProvider)
	systemHandler := systemHandlerPkg.NewSystemHandler(systemService)
	smmHandler := marketingHandlerPkg.NewSMMHandler(socialService)
	mappingService := communicationServicePkg.NewWebhookMappingService(whatsappRepo, messagesService, invoiceService, settingsRepo, lineItemRepo, settingsProvider, orderRepo)

	// Handlers
	orderHandler := orderHandlerPkg.NewOrderHandler(orderService, invoiceService, mappingService)
	syncHandler := syncHandlerPkg.NewSyncHandler(syncService)
	metricsHandler := dashboardHandlerPkg.NewMetricsHandler(metricsService)
	reportHandler := gstHandlerPkg.NewGSTHandler(reportService)
	webhookHandler := webhookHandlerPkg.NewWebhookHandler(webhookService, mappingService, settingsProvider)
	automationHandler := communicationHandlerPkg.NewAutomationHandler(whatsappService, messagesService, mappingService, orderService, customerService, settingsProvider, agentService)
	settingsHandler := configHandlerPkg.NewSettingsHandler(settingsRepo)
	configsHandler := configHandlerPkg.NewConfigsHandler(configsRepo, db)
	redirectHandler := orderHandlerPkg.NewRedirectHandler(orderRepo)
	authHandler := userHandlerPkg.NewAuthHandler(authService)
	customerHandler := orderHandlerPkg.NewCustomerHandler(customerService)
	feedbackHandler := feedbackHandlerPkg.NewFeedbackHandler(feedbackService, settingsProvider, mappingService, whatsappRepo)

	userHandler := userHandlerPkg.NewUserHandler(userService)
	plannerHandler := plannerHandlerPkg.NewPlannerHandler(plannerService, agentService)
	ticketHandler := supportHandlerPkg.NewTicketHandler(ticketService)
	inventoryHandler := inventoryHandlerPkg.NewInventoryHandler(inventoryService)
	oilHandler := productionHandlerPkg.NewOilInventoryHandler(oilService)
	supplierHandler := productionHandlerPkg.NewSupplierHandler(supplierService)
	poHandler := productionHandlerPkg.NewPurchaseOrderHandler(poService)
	mfgHandler := productionHandlerPkg.NewManufacturingHandler(mfgService)
	aiHandler := aiHandlerPkg.NewAIHandler(aiService)
	b2bHandler := b2bHandlerPkg.NewB2BHandler(b2bService)

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
		ticketHandler,
		feedbackHandler,
		inventoryHandler,
		oilHandler,
		supplierHandler,
		poHandler,
		mfgHandler,
		aiHandler,
		b2bHandler,
		authService,
	)

	feedbackScheduler := feedbackServicePkg.NewFeedbackScheduler(settingsProvider, feedbackService, mappingService, whatsappRepo)

	return &Server{
		port:              cfg.Port,
		mux:               mux,
		db:                db,
		amzPoller:         amazonOrderPoller,
		feedbackScheduler: feedbackScheduler,
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
	if s.feedbackScheduler != nil {
		go s.feedbackScheduler.Start(context.Background())
	}

	return server.ListenAndServe()
}
