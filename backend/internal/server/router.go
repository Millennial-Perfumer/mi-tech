package server

import (
	"encoding/json"
	"log"
	"net/http"

	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/handler"
	"mi-tech/internal/service"

	_ "mi-tech/docs"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterRoutes sets up all API routes in one place.
func RegisterRoutes(
	mux *http.ServeMux,
	orderHandler *handler.OrderHandler,
	syncHandler *handler.SyncHandler,
	metricsHandler *handler.MetricsHandler,
	reportHandler *handler.ReportHandler,
	webhookHandler *handler.WebhookHandler,
	automationHandler *whatsapp.AutomationHandler,
	settingsHandler *handler.SettingsHandler,
	configsHandler *handler.ConfigsHandler,
	redirectHandler *handler.RedirectHandler,
	authHandler *handler.AuthHandler,
	customerHandler *handler.CustomerHandler,
	userHandler *handler.UserHandler,
	marketingHandler *handler.MarketingHandler,
	marketingWebhookHandler *handler.MarketingWebhookHandler,
	systemHandler *handler.SystemHandler,
	smmHandler *handler.SMMHandler,
	plannerHandler *handler.PlannerHandler,
	feedbackHandler *handler.FeedbackHandler,
	inventoryHandler *handler.InventoryHandler,
	authService *service.AuthService,
) {
	log.Println("DEBUG: Registering API Routes...")
	cors := CORSMiddleware
	auth := AuthMiddleware(authService)
	metrics := MetricsMiddleware

	// Helper to wrap handlers with both CORS, Auth, and RequireRole("admin")
	adminProtected := func(h http.HandlerFunc) http.HandlerFunc {
		return cors(auth(RequireRole("admin")(h)).ServeHTTP)
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
	mux.HandleFunc("/api/marketing/meta/webhook", marketingWebhookHandler.MetaWebhook)

	// Social Media Management (SMM) Routes
	mux.HandleFunc("/api/marketing/smm/overview", protected(smmHandler.GetOverview))
	mux.HandleFunc("/api/marketing/smm/health", protected(smmHandler.CheckHealth))
	mux.HandleFunc("/api/marketing/smm/post", protected(smmHandler.PostContent))
	mux.HandleFunc("/api/marketing/smm/sync", protected(smmHandler.Sync))
	mux.HandleFunc("/api/marketing/smm/post/insights", protected(smmHandler.GetPostInsights))

	log.Println("DEBUG: Marketing & SMM Routes Registered")

	// Metrics endpoint (unprotected for scraping, but could be internal-only)
	mux.Handle("/api/metrics", promhttp.Handler())

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
	mux.HandleFunc("/api/feedback/submit", feedbackHandler.SubmitFeedback)
	mux.HandleFunc("/api/feedback/validate", feedbackHandler.ValidateFeedback)
	mux.HandleFunc("/api/feedback/config-status", protected(feedbackHandler.GetConfigStatus))
	mux.HandleFunc("/api/feedback", protected(feedbackHandler.GetFeedback))

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
	mux.HandleFunc("/api/orders/delivered", protected(orderHandler.MarkAsDelivered))
	mux.HandleFunc("/api/feedback/scan", protected(feedbackHandler.ScanFeedbackCandidates))
	mux.HandleFunc("/api/feedback/bulk-send", protected(feedbackHandler.BulkSendFeedbackRequests))
	mux.HandleFunc("/api/orders/feedback", protected(orderHandler.GetFeedback))
	mux.HandleFunc("/api/orders/invoice", protected(orderHandler.GenerateInvoice))
	mux.HandleFunc("/api/sources", protected(orderHandler.GetSources))

	// --- Customer Routes ---
	mux.HandleFunc("/api/customers/import", adminProtected(customerHandler.ImportCSV))
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

	mux.HandleFunc("/api/customers/", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
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

	// --- Report Routes ---
	mux.HandleFunc("/api/reports/summary", protected(reportHandler.GetGSTSummary))
	mux.HandleFunc("/api/reports/state-wise", protected(reportHandler.GetStateSummary))
	mux.HandleFunc("/api/reports/hsn-wise", protected(reportHandler.GetHSNSummary))
	mux.HandleFunc("/api/reports/documents-issued", protected(reportHandler.GetDocumentsIssued))

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
	mux.HandleFunc("/api/automation/whatsapp/triggers", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(automationHandler.CreateTrigger)(w, r)
		case http.MethodPut:
			adminProtected(automationHandler.UpdateTrigger)(w, r)
		case http.MethodDelete:
			adminProtected(automationHandler.DeleteTrigger)(w, r)
		default:
			automationHandler.GetTriggers(w, r)
		}
	}))
	mux.HandleFunc("/api/automation/whatsapp/messages", protected(automationHandler.GetMessages))
	mux.HandleFunc("/api/automation/whatsapp/messages/order", protected(automationHandler.GetOrderMessages))

	mux.HandleFunc("/api/automation/whatsapp/conversations", protected(automationHandler.GetConversations))
	mux.HandleFunc("/api/automation/whatsapp/chat", protected(automationHandler.GetChatMessages))
	mux.HandleFunc("/api/automation/whatsapp/chat/upload", adminProtected(automationHandler.UploadChatMedia))
	mux.HandleFunc("/api/automation/whatsapp/chat/send-media", adminProtected(automationHandler.SendChatMedia))
	mux.HandleFunc("/api/automation/whatsapp/send-message", adminProtected(automationHandler.SendFreeTextMessage))
	mux.HandleFunc("/api/automation/whatsapp/conversations/mode", adminProtected(automationHandler.UpdateConversationMode))
	mux.HandleFunc("/api/automation/whatsapp/events", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(automationHandler.CreateEvent)(w, r)
		case http.MethodDelete:
			adminProtected(automationHandler.DeleteEvent)(w, r)
		default:
			automationHandler.GetEvents(w, r)
		}
	}))
	mux.HandleFunc("/api/automation/whatsapp/send-manual", adminProtected(automationHandler.SendManualMessage))
	mux.HandleFunc("/api/automation/whatsapp/send-bulk", adminProtected(automationHandler.SendBulkMarketing))
	mux.HandleFunc("/api/automation/whatsapp/sync-metrics", adminProtected(automationHandler.SyncAutomationMetrics))
	mux.HandleFunc("/api/automation/whatsapp/webhook", automationHandler.WhatsAppWebhook)
	mux.HandleFunc("/api/automation/whatsapp/telegram-webhook", automationHandler.TelegramWebhook)
	mux.HandleFunc("/api/automation/whatsapp/media", protected(automationHandler.GetWhatsAppMedia))
	// Marketing routes moved to top

	// --- Swagger ---
	mux.Handle("/api/swagger/", httpSwagger.WrapHandler)

	// --- Redirect Tracking ---
	mux.HandleFunc("/t/", redirectHandler.RedirectTracking)

	// --- Knowledge API (System Docs) ---
	mux.HandleFunc("/api/system/docs", protected(systemHandler.ListDocs))
	mux.HandleFunc("/api/system/docs/", protected(systemHandler.GetDoc))

	// --- Planner Routes ---
	mux.HandleFunc("/api/planner/boards", protected(plannerHandler.GetBoards))
	mux.HandleFunc("/api/planner/tasks", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(plannerHandler.CreateTask)(w, r)
		case http.MethodPut:
			adminProtected(plannerHandler.UpdateTask)(w, r)
		case http.MethodDelete:
			adminProtected(plannerHandler.DeleteTask)(w, r)
		default:
			plannerHandler.GetTasks(w, r)
		}
	}))
	mux.HandleFunc("/api/planner/tasks/move", adminProtected(plannerHandler.MoveTask))
	mux.HandleFunc("/api/planner/sprints", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(plannerHandler.CreateSprint)(w, r)
		case http.MethodPut:
			adminProtected(plannerHandler.UpdateSprint)(w, r)
		case http.MethodDelete:
			adminProtected(plannerHandler.DeleteSprint)(w, r)
		default:
			plannerHandler.GetSprints(w, r)
		}
	}))
	mux.HandleFunc("/api/planner/analytics", protected(plannerHandler.GetAnalytics))
	
	// --- Inventory Hub Routes ---
	mux.HandleFunc("/api/inventory", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			adminProtected(inventoryHandler.CreateItem)(w, r)
		default:
			inventoryHandler.GetDashboard(w, r)
		}
	}))
	mux.HandleFunc("/api/inventory/next-sku", protected(inventoryHandler.GetNextSKU))
	mux.HandleFunc("/api/inventory/sync-shopify", adminProtected(inventoryHandler.SyncShopify))
	mux.HandleFunc("/api/inventory/map", adminProtected(inventoryHandler.CreateMapping))
	mux.HandleFunc("/api/inventory/adjust", adminProtected(inventoryHandler.AdjustStock))
}
