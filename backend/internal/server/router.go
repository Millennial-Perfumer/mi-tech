package server

import (
	"encoding/json"
	"net/http"

	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/handler"
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
) {
	cors := CORSMiddleware

	// Health check
	mux.HandleFunc("/api/health", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "mi-tech API is running",
		})
	}))

	// --- Order Routes ---
	mux.HandleFunc("/api/orders", cors(orderHandler.GetOrders))
	mux.HandleFunc("/api/orders/status", cors(orderHandler.UpdateOrderStatus))
	mux.HandleFunc("/api/orders/invoice", cors(orderHandler.GenerateInvoice))

	// --- Sync Routes ---
	mux.HandleFunc("/api/shopify/sync", cors(syncHandler.SyncOrders))
	mux.HandleFunc("/api/shopify/reset", cors(syncHandler.ResetOrders))

	// --- Dashboard Metrics ---
	mux.HandleFunc("/api/dashboard/metrics", cors(metricsHandler.GetDashboardMetrics))

	// --- Report Routes ---
	mux.HandleFunc("/api/reports/summary", cors(reportHandler.GetGSTSummary))
	mux.HandleFunc("/api/reports/state-wise", cors(reportHandler.GetStateSummary))
	mux.HandleFunc("/api/reports/hsn-wise", cors(reportHandler.GetHSNSummary))
	mux.HandleFunc("/api/reports/documents-issued", cors(reportHandler.GetDocumentsIssued))

	// --- Webhook Routes ---
	mux.HandleFunc("/api/webhooks/shopify", webhookHandler.ShopifyWebhookHandler)
	mux.HandleFunc("/api/webhook/status", cors(webhookHandler.GetWebhookStatus))

	// --- Settings Routes ---
	mux.HandleFunc("/api/settings/date-range", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			settingsHandler.SetDateRange(w, r)
		default:
			settingsHandler.GetDateRange(w, r)
		}
	}))

	// --- WhatsApp Automation Routes ---
	mux.HandleFunc("/api/automation/whatsapp/metrics", cors(automationHandler.GetAutomationMetrics))
	mux.HandleFunc("/api/automation/whatsapp/templates/upload", cors(automationHandler.UploadTemplateMedia))
	mux.HandleFunc("/api/automation/whatsapp/templates", cors(func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/api/automation/whatsapp/triggers", cors(func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/api/automation/whatsapp/messages", cors(automationHandler.GetMessages))
	mux.HandleFunc("/api/automation/whatsapp/webhook", automationHandler.WhatsAppWebhook)

	// --- Redirect Tracking ---
	mux.HandleFunc("/t/", redirectHandler.RedirectTracking)
}
