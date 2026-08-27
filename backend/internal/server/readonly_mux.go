package server

import (
	"encoding/json"
	"net/http"

	abandonedCheckoutHandlerPkg "mi-tech/internal/domain/abandoned_checkout/handler"
	aiHandlerPkg "mi-tech/internal/domain/ai/handler"
	b2bHandlerPkg "mi-tech/internal/domain/b2b/handler"
	communicationHandlerPkg "mi-tech/internal/domain/communication/handler"
	dashboardHandlerPkg "mi-tech/internal/domain/dashboard/handler"
	feedbackHandlerPkg "mi-tech/internal/domain/feedback/handler"
	gstHandlerPkg "mi-tech/internal/domain/gst/handler"
	inventoryHandlerPkg "mi-tech/internal/domain/inventory/handler"
	marketingHandlerPkg "mi-tech/internal/domain/marketing/handler"
	orderHandlerPkg "mi-tech/internal/domain/order/handler"
	plannerHandlerPkg "mi-tech/internal/domain/planner/handler"
	productionHandlerPkg "mi-tech/internal/domain/production/handler"
	supportHandlerPkg "mi-tech/internal/domain/support/handler"
	syncHandlerPkg "mi-tech/internal/domain/sync/handler"
	configHandlerPkg "mi-tech/internal/shared/config/handler"
	systemHandlerPkg "mi-tech/internal/shared/system/handler"
)

// readOnlyHandlers bundles the handler references needed by the read-only MCP
// mux. It mirrors the subset of handlers exposed via GET-only routes.
type readOnlyHandlers struct {
	orderHandler      *orderHandlerPkg.OrderHandler
	customerHandler   *orderHandlerPkg.CustomerHandler
	metricsHandler    *dashboardHandlerPkg.MetricsHandler
	reportHandler     *gstHandlerPkg.GSTHandler
	inventoryHandler  *inventoryHandlerPkg.InventoryHandler
	oilHandler        *productionHandlerPkg.OilInventoryHandler
	supplierHandler   *productionHandlerPkg.SupplierHandler
	poHandler         *productionHandlerPkg.PurchaseOrderHandler
	mfgHandler        *productionHandlerPkg.ManufacturingHandler
	b2bHandler        *b2bHandlerPkg.B2BHandler
	automationHandler *communicationHandlerPkg.AutomationHandler
	marketingHandler  *marketingHandlerPkg.MarketingHandler
	smmHandler        *marketingHandlerPkg.SMMHandler
	judgeMeHandler    *marketingHandlerPkg.JudgeMeHandler
	feedbackHandler   *feedbackHandlerPkg.FeedbackHandler
	acHandler         *abandonedCheckoutHandlerPkg.AbandonedCheckoutHandler
	plannerHandler    *plannerHandlerPkg.PlannerHandler
	ticketHandler     *supportHandlerPkg.TicketHandler
	aiHandler         *aiHandlerPkg.AIHandler
	settingsHandler   *configHandlerPkg.SettingsHandler
	systemHandler     *systemHandlerPkg.SystemHandler
	syncHandler       *syncHandlerPkg.SyncHandler
}

// ro wraps a read-only handler so any non-GET request is rejected at the mux
// boundary (defense in depth; dispatch always sends GET).
func ro(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		fn(w, r)
	}
}

// registerReadOnlyRoutes builds the internal GET-only mux used by the MCP
// server. It registers exactly the read-only paths from the MCP catalog and
// nothing else, guaranteeing the MCP surface is read-only regardless of tool
// definitions.
func registerReadOnlyRoutes(mux *http.ServeMux, h readOnlyHandlers) {
	// Health
	mux.HandleFunc("/api/health", ro(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "mi-tech API is running"})
	}))

	// Orders
	mux.HandleFunc("/api/orders", ro(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "" {
			h.orderHandler.GetOrder(w, r)
		} else {
			h.orderHandler.GetOrders(w, r)
		}
	}))
	mux.HandleFunc("/api/sources", ro(h.orderHandler.GetSources))

	// Customers
	mux.HandleFunc("/api/customers", ro(h.customerHandler.ListCustomers))

	// Dashboard metrics
	mux.HandleFunc("/api/dashboard/metrics", ro(h.metricsHandler.GetDashboardMetrics))
	mux.HandleFunc("/api/dashboard/top-products", ro(h.metricsHandler.GetTopProducts))
	mux.HandleFunc("/api/dashboard/revenue-trend", ro(h.metricsHandler.GetRevenueTrend))
	mux.HandleFunc("/api/dashboard/geo-distribution", ro(h.metricsHandler.GetGeoDistribution))

	// GST reports
	mux.HandleFunc("/api/reports/summary", ro(h.reportHandler.GetGSTSummary))
	mux.HandleFunc("/api/reports/state-wise", ro(h.reportHandler.GetStateSummary))
	mux.HandleFunc("/api/reports/hsn-wise", ro(h.reportHandler.GetHSNSummary))
	mux.HandleFunc("/api/reports/documents-issued", ro(h.reportHandler.GetDocumentsIssued))
	mux.HandleFunc("/api/reports/gstr1-json", ro(h.reportHandler.GetGSTR1JSON))

	// Inventory
	mux.HandleFunc("/api/inventory", ro(h.inventoryHandler.GetDashboard))
	mux.HandleFunc("/api/inventory/next-sku", ro(h.inventoryHandler.GetNextSKU))

	// Production
	mux.HandleFunc("/api/inventory/suppliers", ro(h.supplierHandler.ListSuppliers))
	mux.HandleFunc("/api/inventory/oil", ro(h.oilHandler.ListOils))
	mux.HandleFunc("/api/inventory/po", ro(h.poHandler.List))
	mux.HandleFunc("/api/inventory/manufacturing", ro(h.mfgHandler.List))

	// B2B billing
	mux.HandleFunc("/api/b2b/customers", ro(h.b2bHandler.HandleCustomers))
	mux.HandleFunc("/api/b2b/invoices", ro(h.b2bHandler.HandleInvoices))
	mux.HandleFunc("/api/b2b/invoices/detail", ro(h.b2bHandler.GetInvoiceByID))
	mux.HandleFunc("/api/b2b/invoices/next-number", ro(h.b2bHandler.GetNextInvoiceNumber))
	mux.HandleFunc("/api/b2b/payment-terms", ro(h.b2bHandler.HandlePaymentTerms))
	mux.HandleFunc("/api/b2b/credit-notes", ro(h.b2bHandler.HandleCreditNotes))
	mux.HandleFunc("/api/b2b/debit-notes", ro(h.b2bHandler.HandleDebitNotes))
	mux.HandleFunc("/api/b2b/customers/ledger", ro(h.b2bHandler.GetCustomerLedger))
	mux.HandleFunc("/api/b2b/customers/outstanding", ro(h.b2bHandler.GetOutstandingAgingReport))
	mux.HandleFunc("/api/b2b/gst-periods", ro(h.b2bHandler.HandleGSTPeriods))
	mux.HandleFunc("/api/b2b/proformas", ro(h.b2bHandler.HandleProformas))
	mux.HandleFunc("/api/b2b/proformas/detail", ro(h.b2bHandler.GetProformaByID))
	mux.HandleFunc("/api/b2b/proformas/next-number", ro(h.b2bHandler.GetNextProformaNumber))
	mux.HandleFunc("/api/b2b/proformas/check-expiry", ro(h.b2bHandler.CheckExpiredProformas))

	// WhatsApp automation
	mux.HandleFunc("/api/automation/whatsapp/metrics", ro(h.automationHandler.GetAutomationMetrics))
	mux.HandleFunc("/api/automation/whatsapp/templates", ro(h.automationHandler.GetTemplates))
	mux.HandleFunc("/api/automation/whatsapp/triggers", ro(h.automationHandler.GetTriggers))
	mux.HandleFunc("/api/automation/whatsapp/messages", ro(h.automationHandler.GetMessages))
	mux.HandleFunc("/api/automation/whatsapp/messages/order", ro(h.automationHandler.GetOrderMessages))
	mux.HandleFunc("/api/automation/whatsapp/conversations", ro(h.automationHandler.GetConversations))
	mux.HandleFunc("/api/automation/whatsapp/chat", ro(h.automationHandler.GetChatMessages))
	mux.HandleFunc("/api/automation/whatsapp/events", ro(h.automationHandler.GetEvents))

	// Marketing (Meta, SMM, Judge.me)
	mux.HandleFunc("/api/marketing/meta/overview", ro(h.marketingHandler.GetMetaOverview))
	mux.HandleFunc("/api/marketing/meta/campaigns", ro(h.marketingHandler.GetMetaCampaigns))
	mux.HandleFunc("/api/marketing/meta/adsets", ro(h.marketingHandler.GetMetaAdSets))
	mux.HandleFunc("/api/marketing/meta/ads", ro(h.marketingHandler.GetMetaAds))
	mux.HandleFunc("/api/marketing/smm/overview", ro(h.smmHandler.GetOverview))
	mux.HandleFunc("/api/marketing/smm/health", ro(h.smmHandler.CheckHealth))
	mux.HandleFunc("/api/marketing/smm/post/insights", ro(h.smmHandler.GetPostInsights))
	mux.HandleFunc("/api/marketing/smm/queue", ro(h.smmHandler.GetQueue))
	mux.HandleFunc("/api/marketing/judgeme/published", ro(h.judgeMeHandler.GetPublishedReviews))

	// Feedback
	mux.HandleFunc("/api/feedback", ro(h.feedbackHandler.GetFeedback))
	mux.HandleFunc("/api/feedback/config-status", ro(h.feedbackHandler.GetConfigStatus))

	// Abandoned checkouts
	mux.HandleFunc("/api/abandoned-checkouts", ro(h.acHandler.GetAbandonedCheckouts))
	mux.HandleFunc("/api/abandoned-checkouts/analytics", ro(h.acHandler.GetAbandonedCheckoutAnalytics))

	// Planner
	mux.HandleFunc("/api/planner/boards", ro(h.plannerHandler.GetBoards))
	mux.HandleFunc("/api/planner/tasks", ro(h.plannerHandler.GetTasks))
	mux.HandleFunc("/api/planner/sprints", ro(h.plannerHandler.GetSprints))
	mux.HandleFunc("/api/planner/analytics", ro(h.plannerHandler.GetAnalytics))

	// Support
	mux.HandleFunc("/api/support/tickets", ro(h.ticketHandler.HandleTickets))

	// AI
	mux.HandleFunc("/api/ai/conversations", ro(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "" {
			h.aiHandler.GetConversation(w, r)
		} else {
			h.aiHandler.ListConversations(w, r)
		}
	}))

	// Settings (masked)
	mux.HandleFunc("/api/settings", ro(h.settingsHandler.GetAllSettings))
	mux.HandleFunc("/api/settings/date-range", ro(h.settingsHandler.GetDateRange))

	// System docs
	mux.HandleFunc("/api/system/docs", ro(h.systemHandler.ListDocs))
	mux.HandleFunc("/api/system/docs/", ro(h.systemHandler.GetDoc))
}

// registerMCPWriteRoutes exposes only the mutation endpoints that are
// explicitly represented in the MCP catalog. It intentionally calls domain
// handlers directly: the MCP machine-key scope is the authorization boundary,
// while the normal browser router remains protected by user JWT/admin checks.
func registerMCPWriteRoutes(mux *http.ServeMux, h readOnlyHandlers) {
	// Orders
	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.orderHandler.CreateOrder(w, r)
		case http.MethodPut:
			h.orderHandler.UpdateOrder(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/orders/status", h.orderHandler.UpdateOrderStatus)
	mux.HandleFunc("/api/orders/payment-status", h.orderHandler.UpdatePaymentStatus)
	mux.HandleFunc("/api/orders/delivered", h.orderHandler.MarkAsDelivered)

	// Customers
	mux.HandleFunc("/api/customers", h.customerHandler.CreateCustomer)
	mux.HandleFunc("/api/customers/bulk-delete", h.customerHandler.BulkDeleteCustomers)
	mux.HandleFunc("/api/customers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.customerHandler.UpdateCustomer(w, r)
		case http.MethodDelete:
			h.customerHandler.DeleteCustomer(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Product inventory
	mux.HandleFunc("/api/inventory", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.inventoryHandler.CreateItem(w, r)
		case http.MethodDelete:
			h.inventoryHandler.Clear(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/inventory/bulk", h.inventoryHandler.BulkCreate)
	mux.HandleFunc("/api/inventory/item", h.inventoryHandler.UpdateItem)
	mux.HandleFunc("/api/inventory/stock", h.inventoryHandler.UpdateStock)
	mux.HandleFunc("/api/inventory/adjust", h.inventoryHandler.AdjustStock)
	mux.HandleFunc("/api/inventory/map", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.inventoryHandler.CreateMapping(w, r)
		} else if r.Method == http.MethodDelete {
			h.inventoryHandler.DeleteMapping(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/inventory/sync-shopify", h.inventoryHandler.SyncShopify)
	mux.HandleFunc("/api/inventory/sync-prices", h.inventoryHandler.SyncPrices)
	mux.HandleFunc("/api/inventory/amazon/sync", h.inventoryHandler.SyncAmazon)

	// Oils, suppliers, purchase orders, and manufacturing
	mux.HandleFunc("/api/inventory/oil", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.oilHandler.CreateOil(w, r)
		case http.MethodPut:
			h.oilHandler.UpdateOil(w, r)
		case http.MethodDelete:
			h.oilHandler.DeleteOil(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/inventory/oil/bulk-delete", h.oilHandler.BulkDeleteOils)
	mux.HandleFunc("/api/inventory/suppliers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.supplierHandler.CreateSupplier(w, r)
		case http.MethodPut:
			h.supplierHandler.UpdateSupplier(w, r)
		case http.MethodDelete:
			h.supplierHandler.DeleteSupplier(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/inventory/po", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.poHandler.Create(w, r)
		case http.MethodPut:
			h.poHandler.Update(w, r)
		case http.MethodDelete:
			h.poHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/inventory/po/bulk", h.poHandler.BulkCreate)
	mux.HandleFunc("/api/inventory/manufacturing", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.mfgHandler.Create(w, r)
		case http.MethodPut:
			h.mfgHandler.Update(w, r)
		case http.MethodDelete:
			h.mfgHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Planner
	mux.HandleFunc("/api/planner/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.plannerHandler.CreateTask(w, r)
		case http.MethodPut:
			h.plannerHandler.UpdateTask(w, r)
		case http.MethodDelete:
			h.plannerHandler.DeleteTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/planner/tasks/move", h.plannerHandler.MoveTask)
	mux.HandleFunc("/api/planner/sprints", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.plannerHandler.CreateSprint(w, r)
		case http.MethodPut:
			h.plannerHandler.UpdateSprint(w, r)
		case http.MethodDelete:
			h.plannerHandler.DeleteSprint(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// B2B billing and proformas
	mux.HandleFunc("/api/b2b/customers", h.b2bHandler.HandleCustomers)
	mux.HandleFunc("/api/b2b/invoices", h.b2bHandler.HandleInvoices)
	mux.HandleFunc("/api/b2b/invoices/issue", h.b2bHandler.IssueInvoice)
	mux.HandleFunc("/api/b2b/invoices/cancel", h.b2bHandler.CancelInvoice)
	mux.HandleFunc("/api/b2b/invoices/deduct-inventory", h.b2bHandler.DeductInventory)
	mux.HandleFunc("/api/b2b/invoices/revert-inventory", h.b2bHandler.RevertInventory)
	mux.HandleFunc("/api/b2b/invoices/payment", h.b2bHandler.UpdatePayment)
	mux.HandleFunc("/api/b2b/payment-terms", h.b2bHandler.HandlePaymentTerms)
	mux.HandleFunc("/api/b2b/credit-notes", h.b2bHandler.HandleCreditNotes)
	mux.HandleFunc("/api/b2b/credit-notes/issue", h.b2bHandler.IssueCreditNote)
	mux.HandleFunc("/api/b2b/credit-notes/cancel", h.b2bHandler.CancelCreditNote)
	mux.HandleFunc("/api/b2b/debit-notes", h.b2bHandler.HandleDebitNotes)
	mux.HandleFunc("/api/b2b/debit-notes/issue", h.b2bHandler.IssueDebitNote)
	mux.HandleFunc("/api/b2b/debit-notes/cancel", h.b2bHandler.CancelDebitNote)
	mux.HandleFunc("/api/b2b/proformas", h.b2bHandler.HandleProformas)
	mux.HandleFunc("/api/b2b/proformas/issue", h.b2bHandler.IssueProforma)
	mux.HandleFunc("/api/b2b/proformas/accept", h.b2bHandler.AcceptProforma)
	mux.HandleFunc("/api/b2b/proformas/reject", h.b2bHandler.RejectProforma)
	mux.HandleFunc("/api/b2b/proformas/cancel", h.b2bHandler.CancelProforma)
	mux.HandleFunc("/api/b2b/proformas/revision", h.b2bHandler.CreateRevision)
	mux.HandleFunc("/api/b2b/proformas/convert", h.b2bHandler.ConvertToTaxInvoice)
	mux.HandleFunc("/api/b2b/proformas/check-expiry", h.b2bHandler.CheckExpiredProformas)

	// Synchronization and settings
	mux.HandleFunc("/api/shopify/sync", h.syncHandler.SyncOrders)
	mux.HandleFunc("/api/shopify/reset", h.syncHandler.ResetOrders)
	mux.HandleFunc("/api/settings/date-range", h.settingsHandler.SetDateRange)

	// Support and abandoned checkouts
	mux.HandleFunc("/api/support/tickets", h.ticketHandler.HandleTickets)
	mux.HandleFunc("/api/support/tickets/", h.ticketHandler.UpdateTicketStatus)
	mux.HandleFunc("/api/abandoned-checkouts", h.acHandler.GetAbandonedCheckouts)
	mux.HandleFunc("/api/abandoned-checkouts/recover", h.acHandler.RecoverCheckout)
	mux.HandleFunc("/api/abandoned-checkouts/status", h.acHandler.UpdateCheckoutStatus)

	// WhatsApp automation
	mux.HandleFunc("/api/automation/whatsapp/templates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.automationHandler.CreateTemplate(w, r)
		case http.MethodPut:
			h.automationHandler.UpdateTemplate(w, r)
		case http.MethodDelete:
			h.automationHandler.DeleteTemplate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/automation/whatsapp/templates/sync", h.automationHandler.SyncTemplateStatus)
	mux.HandleFunc("/api/automation/whatsapp/templates/sync-all", h.automationHandler.SyncAllTemplates)
	mux.HandleFunc("/api/automation/whatsapp/templates/sync-single", h.automationHandler.SyncSingleTemplate)
	mux.HandleFunc("/api/automation/whatsapp/templates/fetch", h.automationHandler.FetchTemplateFromMeta)
	mux.HandleFunc("/api/automation/whatsapp/triggers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.automationHandler.CreateTrigger(w, r)
		case http.MethodPut:
			h.automationHandler.UpdateTrigger(w, r)
		case http.MethodDelete:
			h.automationHandler.DeleteTrigger(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/automation/whatsapp/send-message", h.automationHandler.SendFreeTextMessage)
	mux.HandleFunc("/api/automation/whatsapp/send-manual", h.automationHandler.SendManualMessage)
	mux.HandleFunc("/api/automation/whatsapp/send-bulk", h.automationHandler.SendBulkMarketing)
	mux.HandleFunc("/api/automation/whatsapp/conversations/mode", h.automationHandler.UpdateConversationMode)
	mux.HandleFunc("/api/automation/whatsapp/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.automationHandler.CreateEvent(w, r)
		case http.MethodDelete:
			h.automationHandler.DeleteEvent(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/automation/whatsapp/sync-metrics", h.automationHandler.SyncAutomationMetrics)

	// Social marketing and reviews
	mux.HandleFunc("/api/marketing/smm/post", h.smmHandler.PostContent)
	mux.HandleFunc("/api/marketing/smm/sync", h.smmHandler.Sync)
	mux.HandleFunc("/api/marketing/judgeme/generate", h.judgeMeHandler.GenerateReviews)
	mux.HandleFunc("/api/marketing/judgeme/submit", h.judgeMeHandler.SubmitReviews)

	// Feedback and AI
	mux.HandleFunc("/api/feedback/bulk-send", h.feedbackHandler.BulkSendFeedbackRequests)
	mux.HandleFunc("/api/orders/feedback/comment", h.feedbackHandler.UpdateFeedbackAdminComment)
	mux.HandleFunc("/api/orders/feedback/post-judgeme", h.feedbackHandler.PostJudgeMeReview)
	mux.HandleFunc("/api/orders/feedback/request-google-review", h.feedbackHandler.RequestGoogleReview)
	mux.HandleFunc("/api/ai/chat", h.aiHandler.Chat)
	mux.HandleFunc("/api/ai/conversations", h.aiHandler.DeleteConversation)
}
