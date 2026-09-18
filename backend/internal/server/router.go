package server

import (
	"encoding/json"
	"log"
	"net/http"

	abandonedCheckoutHandlerPkg "mi-tech/internal/domain/abandoned_checkout/handler"
	amazonHandlerPkg "mi-tech/internal/domain/amazon/handler"
	b2bHandlerPkg "mi-tech/internal/domain/b2b/handler"
	communicationHandlerPkg "mi-tech/internal/domain/communication/handler"
	dashboardHandlerPkg "mi-tech/internal/domain/dashboard/handler"
	feedbackHandlerPkg "mi-tech/internal/domain/feedback/handler"
	gstHandlerPkg "mi-tech/internal/domain/gst/handler"
	inventoryHandlerPkg "mi-tech/internal/domain/inventory/handler"
	marketingHandlerPkg "mi-tech/internal/domain/marketing/handler"
	orderHandlerPkg "mi-tech/internal/domain/order/handler"
	productionHandlerPkg "mi-tech/internal/domain/production/handler"
	smmQueueHandlerPkg "mi-tech/internal/domain/smm_queue/handler"
	supportHandlerPkg "mi-tech/internal/domain/support/handler"
	syncHandlerPkg "mi-tech/internal/domain/sync/handler"
	userHandlerPkg "mi-tech/internal/domain/user/handler"
	userServicePkg "mi-tech/internal/domain/user/service"
	webhookHandlerPkg "mi-tech/internal/domain/webhook/handler"
	mcpHandlerPkg "mi-tech/internal/mcp/handler"
	configHandlerPkg "mi-tech/internal/shared/config/handler"
	"mi-tech/internal/shared/middleware"
	systemHandlerPkg "mi-tech/internal/shared/system/handler"

	_ "mi-tech/docs"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterRoutes sets up all API routes in one place.
func RegisterRoutes(
	mux *http.ServeMux,
	orderHandler *orderHandlerPkg.OrderHandler,
	historyHandler *orderHandlerPkg.HistoryHandler,
	syncHandler *syncHandlerPkg.SyncHandler,
	metricsHandler *dashboardHandlerPkg.MetricsHandler,
	reportHandler *gstHandlerPkg.GSTHandler,
	webhookHandler *webhookHandlerPkg.WebhookHandler,
	automationHandler *communicationHandlerPkg.AutomationHandler,
	settingsHandler *configHandlerPkg.SettingsHandler,
	configsHandler *configHandlerPkg.ConfigsHandler,
	redirectHandler *orderHandlerPkg.RedirectHandler,
	authHandler *userHandlerPkg.AuthHandler,
	customerHandler *orderHandlerPkg.CustomerHandler,
	userHandler *userHandlerPkg.UserHandler,
	marketingHandler *marketingHandlerPkg.MarketingHandler,
	marketingWebhookHandler *marketingHandlerPkg.MarketingWebhookHandler,
	systemHandler *systemHandlerPkg.SystemHandler,
	ticketHandler *supportHandlerPkg.TicketHandler,
	feedbackHandler *feedbackHandlerPkg.FeedbackHandler,
	inventoryHandler *inventoryHandlerPkg.InventoryHandler,
	amazonHandler *amazonHandlerPkg.AmazonHandler,
	oilHandler *productionHandlerPkg.OilInventoryHandler,
	supplierHandler *productionHandlerPkg.SupplierHandler,
	poHandler *productionHandlerPkg.PurchaseOrderHandler,
	mfgHandler *productionHandlerPkg.ManufacturingHandler,
	b2bHandler *b2bHandlerPkg.B2BHandler,
	acHandler *abandonedCheckoutHandlerPkg.AbandonedCheckoutHandler,
	judgeMeHandler *marketingHandlerPkg.JudgeMeHandler,
	machineKeyHandler *mcpHandlerPkg.MachineKeyHandler,
	authService *userServicePkg.AuthService,
	smmQueueHandler *smmQueueHandlerPkg.SMMQueueHandler,
) {
	log.Println("DEBUG: Registering API Routes...")
	cors := middleware.CORSMiddleware
	auth := middleware.AuthMiddleware(authService)
	metrics := middleware.MetricsMiddleware

	// Helper to wrap handlers with both CORS, Auth, and RequireRole("admin")
	adminProtected := func(h http.HandlerFunc) http.HandlerFunc {
		return cors(auth(middleware.RequireRole("admin")(h)).ServeHTTP)
	}

	// Helper to wrap handlers with both CORS and Auth (for read/admin)
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return cors(auth(h).ServeHTTP)
	}

	// Force-register marketing routes early to prevent potential shadowing
	mux.HandleFunc("/api/marketing/meta/overview", protected(marketingHandler.GetMetaOverview))
	mux.HandleFunc("/api/marketing/meta/campaigns", protected(marketingHandler.GetMetaCampaigns))
	mux.HandleFunc("/api/marketing/meta/adsets", protected(marketingHandler.GetMetaAdSets))
	mux.HandleFunc("/api/marketing/meta/ads", protected(marketingHandler.GetMetaAds))
	mux.HandleFunc("/api/marketing/meta/webhook", metrics(cors(marketingWebhookHandler.MetaWebhook)).ServeHTTP)

	// Judge.me Review Generator Routes
	if judgeMeHandler != nil {
		mux.HandleFunc("/api/marketing/judgeme/generate", protected(judgeMeHandler.GenerateReviews))
		mux.HandleFunc("/api/marketing/judgeme/submit", protected(judgeMeHandler.SubmitReviews))
		mux.HandleFunc("/api/marketing/judgeme/published", protected(judgeMeHandler.GetPublishedReviews))
	}

	log.Println("DEBUG: Marketing & Judge.me Routes Registered")

	// Metrics endpoint (unprotected for scraping, but could be internal-only)
	mux.Handle("/api/metrics", cors(promhttp.Handler().ServeHTTP))

	// --- SMM Queue Routes ---
	// Queue submissions are admin-only because they publish content for the
	// connected social media accounts.
	mux.HandleFunc("/api/smm-queue", adminProtected(smmQueueHandler.Handle))
	mux.HandleFunc("/api/smm-queue/", adminProtected(smmQueueHandler.HandleItem))

	// Health check
	mux.HandleFunc("/api/health", metrics(cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "mi-tech API is running",
		})
	})).ServeHTTP)

	// --- Feedback Routes ---
	mux.HandleFunc("/api/feedback/submit", metrics(cors(feedbackHandler.SubmitFeedback)).ServeHTTP)
	mux.HandleFunc("/api/feedback/validate", metrics(cors(feedbackHandler.ValidateFeedback)).ServeHTTP)
	mux.HandleFunc("/api/feedback/config-status", protected(feedbackHandler.GetConfigStatus))
	mux.HandleFunc("/api/feedback", protected(feedbackHandler.GetFeedback))
	mux.HandleFunc("/api/orders/feedback/post-judgeme", protected(feedbackHandler.PostJudgeMeReview))
	mux.HandleFunc("/api/orders/feedback/request-google-review", protected(feedbackHandler.RequestGoogleReview))

	// --- Auth Routes ---
	mux.HandleFunc("/api/auth/login", metrics(cors(authHandler.Login)).ServeHTTP)
	mux.HandleFunc("/api/auth/verify-otp", metrics(cors(authHandler.VerifyOTP)).ServeHTTP)
	mux.HandleFunc("/api/auth/verify", metrics(protected(authHandler.VerifyAuth)).ServeHTTP)

	// --- User Routes ---
	mux.HandleFunc("/api/users", adminProtected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			userHandler.GetUsers(w, r)
		}
	}))

	// --- Order Routes ---
	mux.HandleFunc("/api/orders", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(orderHandler.CreateOrder)(w, r)
		case http.MethodPut:
			adminProtected(orderHandler.UpdateOrder)(w, r)
		default:
			if r.URL.Query().Get("id") != "" {
				orderHandler.GetOrder(w, r)
			} else {
				orderHandler.GetOrders(w, r)
			}
		}
	}))
	mux.HandleFunc("/api/orders/status", protected(orderHandler.UpdateOrderStatus))
	mux.HandleFunc("/api/orders/payment-status", protected(orderHandler.UpdatePaymentStatus))
	mux.HandleFunc("/api/orders/delivered", protected(orderHandler.MarkAsDelivered))
	mux.HandleFunc("/api/orders/history", protected(historyHandler.GetOrderHistory))
	mux.HandleFunc("/api/feedback/scan", protected(feedbackHandler.ScanFeedbackCandidates))
	mux.HandleFunc("/api/feedback/bulk-send", protected(feedbackHandler.BulkSendFeedbackRequests))
	mux.HandleFunc("/api/orders/feedback", protected(feedbackHandler.GetFeedback))
	mux.HandleFunc("/api/orders/feedback/comment", protected(feedbackHandler.UpdateFeedbackAdminComment))
	mux.HandleFunc("/api/orders/invoice", protected(orderHandler.GenerateInvoice))
	mux.HandleFunc("/api/sources", protected(orderHandler.GetSources))

	// --- Customer Routes ---
	mux.HandleFunc("/api/customers/import", adminProtected(customerHandler.ImportCSV))
	mux.HandleFunc("/api/customers/export-meta", protected(customerHandler.ExportMetaCSV))
	mux.HandleFunc("/api/customers/bulk-delete", adminProtected(customerHandler.BulkDeleteCustomers))
	mux.HandleFunc("/api/customers", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(customerHandler.CreateCustomer)(w, r)
		case http.MethodDelete:
			adminProtected(customerHandler.DeleteAllCustomers)(w, r)
		default:
			customerHandler.ListCustomers(w, r)
		}
	}))
	mux.HandleFunc("/api/customers/history", protected(historyHandler.GetCustomerHistory))

	mux.HandleFunc("/api/customers/", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			customerHandler.GetCustomer(w, r)
		case http.MethodPut:
			adminProtected(customerHandler.UpdateCustomer)(w, r)
		case http.MethodDelete:
			adminProtected(customerHandler.DeleteCustomer)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// --- Sync Routes ---
	mux.HandleFunc("/api/shopify/sync", adminProtected(syncHandler.SyncOrders))
	mux.HandleFunc("/api/shopify/reset", adminProtected(syncHandler.ResetOrders))

	// --- Dashboard Metrics ---
	mux.HandleFunc("/api/dashboard/metrics", protected(metricsHandler.GetDashboardMetrics))
	mux.HandleFunc("/api/dashboard/top-products", protected(metricsHandler.GetTopProducts))
	mux.HandleFunc("/api/dashboard/revenue-trend", protected(metricsHandler.GetRevenueTrend))
	mux.HandleFunc("/api/dashboard/geo-distribution", protected(metricsHandler.GetGeoDistribution))

	// --- Report Routes ---
	mux.HandleFunc("/api/reports/summary", protected(reportHandler.GetGSTSummary))
	mux.HandleFunc("/api/reports/state-wise", protected(reportHandler.GetStateSummary))
	mux.HandleFunc("/api/reports/hsn-wise", protected(reportHandler.GetHSNSummary))
	mux.HandleFunc("/api/reports/documents-issued", protected(reportHandler.GetDocumentsIssued))
	mux.HandleFunc("/api/reports/gstr1-json", protected(reportHandler.GetGSTR1JSON))

	// --- Webhook Routes ---
	mux.HandleFunc("/api/webhooks/shopify", webhookHandler.ShopifyWebhookHandler)
	mux.HandleFunc("/api/webhook/status", protected(webhookHandler.GetWebhookStatus))

	// --- Settings Routes ---
	mux.HandleFunc("/api/settings", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			adminProtected(settingsHandler.UpdateSetting)(w, r)
		default:
			settingsHandler.GetAllSettings(w, r)
		}
	}))
	mux.HandleFunc("/api/settings/date-range", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			adminProtected(settingsHandler.SetDateRange)(w, r)
		default:
			settingsHandler.GetDateRange(w, r)
		}
	}))

	// --- Configs Routes (API Keys & Secrets) ---
	mux.HandleFunc("/api/configs", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			adminProtected(configsHandler.UpdateConfig)(w, r)
		default:
			configsHandler.GetAllConfigs(w, r)
		}
	}))
	mux.HandleFunc("/api/configs/reveal", adminProtected(configsHandler.RevealConfigs))

	// --- WhatsApp Automation Routes ---
	mux.HandleFunc("/api/automation/whatsapp/metrics", protected(automationHandler.GetAutomationMetrics))
	mux.HandleFunc("/api/automation/whatsapp/templates/sync", adminProtected(automationHandler.SyncTemplateStatus))
	mux.HandleFunc("/api/automation/whatsapp/templates/sync-all", adminProtected(automationHandler.SyncAllTemplates))
	mux.HandleFunc("/api/automation/whatsapp/templates/sync-single", adminProtected(automationHandler.SyncSingleTemplate))
	mux.HandleFunc("/api/automation/whatsapp/templates/fetch", adminProtected(automationHandler.FetchTemplateFromMeta))
	mux.HandleFunc("/api/automation/whatsapp/templates/upload", adminProtected(automationHandler.UploadTemplateMedia))
	mux.HandleFunc("/api/automation/whatsapp/templates", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(automationHandler.CreateTemplate)(w, r)
		case http.MethodPut:
			adminProtected(automationHandler.UpdateTemplate)(w, r)
		case http.MethodDelete:
			adminProtected(automationHandler.DeleteTemplate)(w, r)
		default:
			automationHandler.GetTemplates(w, r)
		}
	}))
	mux.HandleFunc("/api/automation/whatsapp/triggers", protected(func(w http.ResponseWriter, r *http.Re