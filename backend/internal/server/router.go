package server

import (
	"encoding/json"
	"net/http"

	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/handler"
	"mi-tech/internal/service"
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
	redirectHandler *handler.RedirectHandler,
	authHandler *handler.AuthHandler,
	authService *service.AuthService,
) {
	cors := CORSMiddleware
	auth := AuthMiddleware(authService)

	// Helper to wrap handlers with both CORS and Auth
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return cors(auth(h).ServeHTTP)
	}

	// Health check
	mux.HandleFunc("/api/health", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "mi-tech API is running",
		})
	}))

	// --- Auth Routes ---
	mux.HandleFunc("/api/auth/login", cors(authHandler.Login))

	// --- Order Routes ---
	mux.HandleFunc("/api/orders", protected(orderHandler.GetOrders))
	mux.HandleFunc("/api/orders/status", protected(orderHandler.UpdateOrderStatus))
	mux.HandleFunc("/api/orders/invoice", protected(orderHandler.GenerateInvoice))

	// --- Sync Routes ---
	mux.HandleFunc("/api/shopify/sync", protected(syncHandler.SyncOrders))
	mux.HandleFunc("/api/shopify/reset", protected(syncHandler.ResetOrders))

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
	mux.HandleFunc("/api/settings", protected(settingsHandler.GetAllSettings))
	mux.HandleFunc("/api/settings/date-range", protected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			settingsHandler.SetDateRange(w, r)
		default:
			settingsHandler.GetDateRange(w, r)
		}
	}))

	// --- WhatsApp Automation Routes ---
	mux.HandleFunc("/api/automation/whatsapp/metrics", protected(automationHandler.GetAutomationMetrics))
	mux.HandleFunc("/api/automation/whatsapp/templates/sync", protected(automationHandler.SyncTemplateStatus))
	mux.HandleFunc("/api/automation/whatsapp/templates/upload", protected(automationHandler.UploadTemplateMedia))
	mux.HandleFunc("/api/automation/whatsapp/templates", protected(func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/api/automation/whatsapp/triggers", protected(func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/api/automation/whatsapp/messages", protected(automationHandler.GetMessages))
	mux.HandleFunc("/api/automation/whatsapp/sync-metrics", protected(automationHandler.SyncAutomationMetrics))
	mux.HandleFunc("/api/automation/whatsapp/webhook", automationHandler.WhatsAppWebhook)

	// --- Redirect Tracking ---
	mux.HandleFunc("/t/", redirectHandler.RedirectTracking)
}
