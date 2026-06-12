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

type Router struct {
	mux            *http.ServeMux
	protected      func(http.HandlerFunc) http.HandlerFunc
	adminProtected func(http.HandlerFunc) http.HandlerFunc
	metrics        func(http.Handler) http.Handler
	cors           func(http.HandlerFunc) http.HandlerFunc
}

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
	oilHandler *handler.OilInventoryHandler,
	supplierHandler *handler.SupplierHandler,
	poHandler *handler.PurchaseOrderHandler,
	mfgHandler *handler.ManufacturingHandler,
	aiHandler *handler.AIHandler,
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

	r := &Router{
		mux:            mux,
		protected:      protected,
		adminProtected: adminProtected,
		metrics:        metrics,
		cors:           cors,
	}

	r.registerMarketingRoutes(marketingHandler, marketingWebhookHandler, smmHandler)
	r.registerSystemRoutes(systemHandler, redirectHandler)
	r.registerFeedbackRoutes(feedbackHandler)
	r.registerAuthRoutes(authHandler)
	r.registerUserRoutes(userHandler)
	r.registerOrderRoutes(orderHandler, feedbackHandler)
	r.registerCustomerRoutes(customerHandler)
	r.registerSyncRoutes(syncHandler)
	r.registerMetricsRoutes(metricsHandler)
	r.registerReportRoutes(reportHandler)
	r.registerWebhookRoutes(webhookHandler)
	r.registerSettingsRoutes(settingsHandler)
	r.registerConfigsRoutes(configsHandler)
	r.registerAutomationRoutes(automationHandler)
	r.registerPlannerRoutes(plannerHandler)
	r.registerInventoryRoutes(inventoryHandler)
	r.registerOilRoutes(oilHandler)
	r.registerSupplierRoutes(supplierHandler)
	r.registerPORoutes(poHandler)
	r.registerMfgRoutes(mfgHandler)
	r.registerAIRoutes(aiHandler)
}

func (r *Router) registerMarketingRoutes(marketingHandler *handler.MarketingHandler, marketingWebhookHandler *handler.MarketingWebhookHandler, smmHandler *handler.SMMHandler) {
	r.mux.HandleFunc("/api/marketing/meta/overview", r.protected(marketingHandler.GetMetaOverview))
	r.mux.HandleFunc("/api/marketing/meta/campaigns", r.protected(marketingHandler.GetMetaCampaigns))
	r.mux.HandleFunc("/api/marketing/meta/adsets", r.protected(marketingHandler.GetMetaAdSets))
	r.mux.HandleFunc("/api/marketing/meta/ads", r.protected(marketingHandler.GetMetaAds))
	r.mux.HandleFunc("/api/marketing/meta/webhook", r.metrics(r.cors(marketingWebhookHandler.MetaWebhook)).ServeHTTP)
	r.mux.HandleFunc("/api/marketing/smm/overview", r.protected(smmHandler.GetOverview))
	r.mux.HandleFunc("/api/marketing/smm/health", r.protected(smmHandler.CheckHealth))
	r.mux.HandleFunc("/api/marketing/smm/post", r.protected(smmHandler.PostContent))
	r.mux.HandleFunc("/api/marketing/smm/sync", r.protected(smmHandler.Sync))
	r.mux.HandleFunc("/api/marketing/smm/post/insights", r.protected(smmHandler.GetPostInsights))
}

func (r *Router) registerSystemRoutes(systemHandler *handler.SystemHandler, redirectHandler *handler.RedirectHandler) {
	r.mux.Handle("/api/metrics", r.cors(promhttp.Handler().ServeHTTP))
	r.mux.HandleFunc("/api/health", r.metrics(r.cors(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "mi-tech API is running",
		})
	}))).ServeHTTP)
	r.mux.Handle("/api/swagger/", r.cors(httpSwagger.WrapHandler.ServeHTTP))
	r.mux.HandleFunc("/t/", redirectHandler.RedirectTracking)
	r.mux.HandleFunc("/api/system/docs", r.protected(systemHandler.ListDocs))
	r.mux.HandleFunc("/api/system/docs/", r.protected(systemHandler.GetDoc))
}

func (r *Router) registerFeedbackRoutes(feedbackHandler *handler.FeedbackHandler) {
	r.mux.HandleFunc("/api/feedback/submit", r.metrics(r.cors(feedbackHandler.SubmitFeedback)).ServeHTTP)
	r.mux.HandleFunc("/api/feedback/validate", r.metrics(r.cors(feedbackHandler.ValidateFeedback)).ServeHTTP)
	r.mux.HandleFunc("/api/feedback/config-status", r.protected(feedbackHandler.GetConfigStatus))
	r.mux.HandleFunc("/api/feedback", r.protected(feedbackHandler.GetFeedback))
}

func (r *Router) registerAuthRoutes(authHandler *handler.AuthHandler) {
	r.mux.HandleFunc("/api/auth/login", r.metrics(r.cors(authHandler.Login)).ServeHTTP)
	r.mux.HandleFunc("/api/auth/verify-otp", r.metrics(r.cors(authHandler.VerifyOTP)).ServeHTTP)
	r.mux.HandleFunc("/api/auth/verify", r.metrics(r.protected(authHandler.VerifyAuth)).ServeHTTP)
}

func (r *Router) registerUserRoutes(userHandler *handler.UserHandler) {
	r.mux.HandleFunc("/api/users", r.adminProtected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			userHandler.CreateUser(w, req)
		default:
			userHandler.GetUsers(w, req)
		}
	}))
}

func (r *Router) registerOrderRoutes(orderHandler *handler.OrderHandler, feedbackHandler *handler.FeedbackHandler) {
	r.mux.HandleFunc("/api/orders", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPut:
			r.adminProtected(orderHandler.UpdateOrder)(w, req)
		default:
			if req.URL.Query().Get("id") != "" {
				orderHandler.GetOrder(w, req)
			} else {
				orderHandler.GetOrders(w, req)
			}
		}
	}))
	r.mux.HandleFunc("/api/orders/status", r.protected(orderHandler.UpdateOrderStatus))
	r.mux.HandleFunc("/api/orders/payment-status", r.protected(orderHandler.UpdatePaymentStatus))
	r.mux.HandleFunc("/api/orders/delivered", r.protected(orderHandler.MarkAsDelivered))
	r.mux.HandleFunc("/api/feedback/scan", r.protected(feedbackHandler.ScanFeedbackCandidates))
	r.mux.HandleFunc("/api/feedback/bulk-send", r.protected(feedbackHandler.BulkSendFeedbackRequests))
	r.mux.HandleFunc("/api/orders/feedback", r.protected(orderHandler.GetFeedback))
	r.mux.HandleFunc("/api/orders/feedback/comment", r.protected(orderHandler.UpdateFeedbackAdminComment))
	r.mux.HandleFunc("/api/orders/invoice", r.protected(orderHandler.GenerateInvoice))
	r.mux.HandleFunc("/api/sources", r.protected(orderHandler.GetSources))
}

func (r *Router) registerCustomerRoutes(customerHandler *handler.CustomerHandler) {
	r.mux.HandleFunc("/api/customers/import", r.adminProtected(customerHandler.ImportCSV))
	r.mux.HandleFunc("/api/customers/export-meta", r.protected(customerHandler.ExportMetaCSV))
	r.mux.HandleFunc("/api/customers/bulk-delete", r.adminProtected(customerHandler.BulkDeleteCustomers))
	r.mux.HandleFunc("/api/customers", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(customerHandler.CreateCustomer)(w, req)
		case http.MethodDelete:
			r.adminProtected(customerHandler.DeleteAllCustomers)(w, req)
		default:
			customerHandler.ListCustomers(w, req)
		}
	}))
	r.mux.HandleFunc("/api/customers/", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPut:
			r.adminProtected(customerHandler.UpdateCustomer)(w, req)
		case http.MethodDelete:
			r.adminProtected(customerHandler.DeleteCustomer)(w, req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

func (r *Router) registerSyncRoutes(syncHandler *handler.SyncHandler) {
	r.mux.HandleFunc("/api/shopify/sync", r.adminProtected(syncHandler.SyncOrders))
	r.mux.HandleFunc("/api/shopify/reset", r.adminProtected(syncHandler.ResetOrders))
}

func (r *Router) registerMetricsRoutes(metricsHandler *handler.MetricsHandler) {
	r.mux.HandleFunc("/api/dashboard/metrics", r.protected(metricsHandler.GetDashboardMetrics))
	r.mux.HandleFunc("/api/dashboard/top-products", r.protected(metricsHandler.GetTopProducts))
	r.mux.HandleFunc("/api/dashboard/revenue-trend", r.protected(metricsHandler.GetRevenueTrend))
	r.mux.HandleFunc("/api/dashboard/geo-distribution", r.protected(metricsHandler.GetGeoDistribution))
}

func (r *Router) registerReportRoutes(reportHandler *handler.ReportHandler) {
	r.mux.HandleFunc("/api/reports/summary", r.protected(reportHandler.GetGSTSummary))
	r.mux.HandleFunc("/api/reports/state-wise", r.protected(reportHandler.GetStateSummary))
	r.mux.HandleFunc("/api/reports/hsn-wise", r.protected(reportHandler.GetHSNSummary))
	r.mux.HandleFunc("/api/reports/documents-issued", r.protected(reportHandler.GetDocumentsIssued))
}

func (r *Router) registerWebhookRoutes(webhookHandler *handler.WebhookHandler) {
	r.mux.HandleFunc("/api/webhooks/shopify", webhookHandler.ShopifyWebhookHandler)
	r.mux.HandleFunc("/api/webhook/status", r.protected(webhookHandler.GetWebhookStatus))
}

func (r *Router) registerSettingsRoutes(settingsHandler *handler.SettingsHandler) {
	r.mux.HandleFunc("/api/settings", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPut:
			r.adminProtected(settingsHandler.UpdateSetting)(w, req)
		default:
			settingsHandler.GetAllSettings(w, req)
		}
	}))
	r.mux.HandleFunc("/api/settings/date-range", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPut:
			r.adminProtected(settingsHandler.SetDateRange)(w, req)
		default:
			settingsHandler.GetDateRange(w, req)
		}
	}))
}

func (r *Router) registerConfigsRoutes(configsHandler *handler.ConfigsHandler) {
	r.mux.HandleFunc("/api/configs", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPut:
			r.adminProtected(configsHandler.UpdateConfig)(w, req)
		default:
			configsHandler.GetAllConfigs(w, req)
		}
	}))
	r.mux.HandleFunc("/api/configs/reveal", r.adminProtected(configsHandler.RevealConfigs))
}

func (r *Router) registerAutomationRoutes(automationHandler *whatsapp.AutomationHandler) {
	r.mux.HandleFunc("/api/automation/whatsapp/metrics", r.protected(automationHandler.GetAutomationMetrics))
	r.mux.HandleFunc("/api/automation/whatsapp/templates/sync", r.adminProtected(automationHandler.SyncTemplateStatus))
	r.mux.HandleFunc("/api/automation/whatsapp/templates/sync-all", r.adminProtected(automationHandler.SyncAllTemplates))
	r.mux.HandleFunc("/api/automation/whatsapp/templates/sync-single", r.adminProtected(automationHandler.SyncSingleTemplate))
	r.mux.HandleFunc("/api/automation/whatsapp/templates/fetch", r.adminProtected(automationHandler.FetchTemplateFromMeta))
	r.mux.HandleFunc("/api/automation/whatsapp/templates/upload", r.adminProtected(automationHandler.UploadTemplateMedia))
	r.mux.HandleFunc("/api/automation/whatsapp/templates", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(automationHandler.CreateTemplate)(w, req)
		case http.MethodPut:
			r.adminProtected(automationHandler.UpdateTemplate)(w, req)
		case http.MethodDelete:
			r.adminProtected(automationHandler.DeleteTemplate)(w, req)
		default:
			automationHandler.GetTemplates(w, req)
		}
	}))
	r.mux.HandleFunc("/api/automation/whatsapp/triggers", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(automationHandler.CreateTrigger)(w, req)
		case http.MethodPut:
			r.adminProtected(automationHandler.UpdateTrigger)(w, req)
		case http.MethodDelete:
			r.adminProtected(automationHandler.DeleteTrigger)(w, req)
		default:
			automationHandler.GetTriggers(w, req)
		}
	}))
	r.mux.HandleFunc("/api/automation/whatsapp/messages", r.protected(automationHandler.GetMessages))
	r.mux.HandleFunc("/api/automation/whatsapp/messages/order", r.protected(automationHandler.GetOrderMessages))
	r.mux.HandleFunc("/api/automation/whatsapp/conversations", r.protected(automationHandler.GetConversations))
	r.mux.HandleFunc("/api/automation/whatsapp/chat", r.protected(automationHandler.GetChatMessages))
	r.mux.HandleFunc("/api/automation/whatsapp/chat/upload", r.adminProtected(automationHandler.UploadChatMedia))
	r.mux.HandleFunc("/api/automation/whatsapp/chat/send-media", r.adminProtected(automationHandler.SendChatMedia))
	r.mux.HandleFunc("/api/automation/whatsapp/send-message", r.adminProtected(automationHandler.SendFreeTextMessage))
	r.mux.HandleFunc("/api/automation/whatsapp/conversations/mode", r.adminProtected(automationHandler.UpdateConversationMode))
	r.mux.HandleFunc("/api/automation/whatsapp/events", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(automationHandler.CreateEvent)(w, req)
		case http.MethodDelete:
			r.adminProtected(automationHandler.DeleteEvent)(w, req)
		default:
			automationHandler.GetEvents(w, req)
		}
	}))
	r.mux.HandleFunc("/api/automation/whatsapp/send-manual", r.adminProtected(automationHandler.SendManualMessage))
	r.mux.HandleFunc("/api/automation/whatsapp/send-bulk", r.adminProtected(automationHandler.SendBulkMarketing))
	r.mux.HandleFunc("/api/automation/whatsapp/sync-metrics", r.adminProtected(automationHandler.SyncAutomationMetrics))
	r.mux.HandleFunc("/api/automation/whatsapp/webhook", automationHandler.WhatsAppWebhook)
	r.mux.HandleFunc("/api/automation/whatsapp/telegram-webhook", automationHandler.TelegramWebhook)
	r.mux.HandleFunc("/api/automation/whatsapp/media", r.protected(automationHandler.GetWhatsAppMedia))
}

func (r *Router) registerPlannerRoutes(plannerHandler *handler.PlannerHandler) {
	r.mux.HandleFunc("/api/planner/boards", r.protected(plannerHandler.GetBoards))
	r.mux.HandleFunc("/api/planner/tasks", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(plannerHandler.CreateTask)(w, req)
		case http.MethodPut:
			r.adminProtected(plannerHandler.UpdateTask)(w, req)
		case http.MethodDelete:
			r.adminProtected(plannerHandler.DeleteTask)(w, req)
		default:
			plannerHandler.GetTasks(w, req)
		}
	}))
	r.mux.HandleFunc("/api/planner/tasks/move", r.adminProtected(plannerHandler.MoveTask))
	r.mux.HandleFunc("/api/planner/sprints", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(plannerHandler.CreateSprint)(w, req)
		case http.MethodPut:
			r.adminProtected(plannerHandler.UpdateSprint)(w, req)
		case http.MethodDelete:
			r.adminProtected(plannerHandler.DeleteSprint)(w, req)
		default:
			plannerHandler.GetSprints(w, req)
		}
	}))
	r.mux.HandleFunc("/api/planner/analytics", r.protected(plannerHandler.GetAnalytics))
}

func (r *Router) registerInventoryRoutes(inventoryHandler *handler.InventoryHandler) {
	r.mux.HandleFunc("/api/inventory", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(inventoryHandler.CreateItem)(w, req)
		case http.MethodDelete:
			r.adminProtected(inventoryHandler.Clear)(w, req)
		default:
			inventoryHandler.GetDashboard(w, req)
		}
	}))
	r.mux.HandleFunc("/api/inventory/next-sku", r.protected(inventoryHandler.GetNextSKU))
	r.mux.HandleFunc("/api/inventory/sync-shopify", r.adminProtected(inventoryHandler.SyncShopify))
	r.mux.HandleFunc("/api/inventory/bulk", r.adminProtected(inventoryHandler.BulkCreate))
	r.mux.HandleFunc("/api/inventory/map", r.adminProtected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			inventoryHandler.CreateMapping(w, req)
		case http.MethodDelete:
			inventoryHandler.DeleteMapping(w, req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	r.mux.HandleFunc("/api/inventory/stock", r.adminProtected(inventoryHandler.UpdateStock))
	r.mux.HandleFunc("/api/inventory/adjust", r.adminProtected(inventoryHandler.AdjustStock))
	r.mux.HandleFunc("/api/inventory/logs", r.adminProtected(inventoryHandler.GetLogs))
	r.mux.HandleFunc("/api/inventory/amazon/sync", r.adminProtected(inventoryHandler.SyncAmazon))
	r.mux.HandleFunc("/api/inventory/item", r.adminProtected(inventoryHandler.UpdateItem))
}

func (r *Router) registerOilRoutes(oilHandler *handler.OilInventoryHandler) {
	r.mux.HandleFunc("/api/inventory/oil/bulk-delete", r.adminProtected(oilHandler.BulkDeleteOils))
	r.mux.HandleFunc("/api/inventory/oil", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(oilHandler.CreateOil)(w, req)
		case http.MethodPut:
			r.adminProtected(oilHandler.UpdateOil)(w, req)
		case http.MethodDelete:
			r.adminProtected(oilHandler.DeleteOil)(w, req)
		default:
			oilHandler.ListOils(w, req)
		}
	}))
}

func (r *Router) registerSupplierRoutes(supplierHandler *handler.SupplierHandler) {
	r.mux.HandleFunc("/api/inventory/suppliers", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(supplierHandler.CreateSupplier)(w, req)
		case http.MethodPut:
			r.adminProtected(supplierHandler.UpdateSupplier)(w, req)
		case http.MethodDelete:
			r.adminProtected(supplierHandler.DeleteSupplier)(w, req)
		default:
			supplierHandler.ListSuppliers(w, req)
		}
	}))
}

func (r *Router) registerPORoutes(poHandler *handler.PurchaseOrderHandler) {
	r.mux.HandleFunc("/api/inventory/po", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(poHandler.Create)(w, req)
		case http.MethodPut:
			r.adminProtected(poHandler.Update)(w, req)
		case http.MethodDelete:
			r.adminProtected(poHandler.Delete)(w, req)
		default:
			poHandler.List(w, req)
		}
	}))
	r.mux.HandleFunc("/api/inventory/po/bulk", r.adminProtected(poHandler.BulkCreate))
}

func (r *Router) registerMfgRoutes(mfgHandler *handler.ManufacturingHandler) {
	r.mux.HandleFunc("/api/inventory/manufacturing", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.adminProtected(mfgHandler.Create)(w, req)
		case http.MethodPut:
			r.adminProtected(mfgHandler.Update)(w, req)
		case http.MethodDelete:
			r.adminProtected(mfgHandler.Delete)(w, req)
		default:
			mfgHandler.List(w, req)
		}
	}))
}

func (r *Router) registerAIRoutes(aiHandler *handler.AIHandler) {
	r.mux.HandleFunc("/api/ai/chat", r.protected(aiHandler.Chat))
	r.mux.HandleFunc("/api/ai/conversations", r.protected(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if req.URL.Query().Get("id") != "" {
				aiHandler.GetConversation(w, req)
			} else {
				aiHandler.ListConversations(w, req)
			}
		case http.MethodDelete:
			aiHandler.DeleteConversation(w, req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}
