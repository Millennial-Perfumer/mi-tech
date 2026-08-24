package mcp

// RouteBinding describes how an MCP tool maps to an internal backend route.
// The internal MCP mux only ever serves these paths via GET, guaranteeing the
// read-only guarantee for the whole catalog.
type RouteBinding struct {
	// Path is the internal route (no query string).
	Path string
	// Method is always http.MethodGet for read-only tools.
	Method string
}

// routeMap is the single source of truth for the read-only route surface.
// It is keyed by tool name and validated against DefaultCatalog in tests.
var routeMap = map[string]RouteBinding{
	// Orders
	"orders_list":    {Path: "/api/orders", Method: "GET"},
	"orders_get":     {Path: "/api/orders", Method: "GET"},
	"orders_sources": {Path: "/api/sources", Method: "GET"},
	// Customers
	"customers_list": {Path: "/api/customers", Method: "GET"},
	// Dashboard metrics
	"dashboard_metrics":          {Path: "/api/dashboard/metrics", Method: "GET"},
	"dashboard_top_products":     {Path: "/api/dashboard/top-products", Method: "GET"},
	"dashboard_revenue_trend":    {Path: "/api/dashboard/revenue-trend", Method: "GET"},
	"dashboard_geo_distribution": {Path: "/api/dashboard/geo-distribution", Method: "GET"},
	// GST reports
	"gst_summary":          {Path: "/api/reports/summary", Method: "GET"},
	"gst_state_wise":       {Path: "/api/reports/state-wise", Method: "GET"},
	"gst_hsn_wise":         {Path: "/api/reports/hsn-wise", Method: "GET"},
	"gst_documents_issued": {Path: "/api/reports/documents-issued", Method: "GET"},
	"gst_gstr1_json":       {Path: "/api/reports/gstr1-json", Method: "GET"},
	// Inventory
	"inventory_dashboard": {Path: "/api/inventory", Method: "GET"},
	"inventory_next_sku":  {Path: "/api/inventory/next-sku", Method: "GET"},
	// Production
	"suppliers_list":       {Path: "/api/inventory/suppliers", Method: "GET"},
	"oils_list":            {Path: "/api/inventory/oil", Method: "GET"},
	"purchase_orders_list": {Path: "/api/inventory/po", Method: "GET"},
	"manufacturing_list":   {Path: "/api/inventory/manufacturing", Method: "GET"},
	// B2B billing
	"b2b_customers_list":         {Path: "/api/b2b/customers", Method: "GET"},
	"b2b_invoices_list":          {Path: "/api/b2b/invoices", Method: "GET"},
	"b2b_invoice_get":            {Path: "/api/b2b/invoices/detail", Method: "GET"},
	"b2b_invoice_next_number":    {Path: "/api/b2b/invoices/next-number", Method: "GET"},
	"b2b_payment_terms_list":     {Path: "/api/b2b/payment-terms", Method: "GET"},
	"b2b_credit_notes_list":      {Path: "/api/b2b/credit-notes", Method: "GET"},
	"b2b_debit_notes_list":       {Path: "/api/b2b/debit-notes", Method: "GET"},
	"b2b_customer_ledger":        {Path: "/api/b2b/customers/ledger", Method: "GET"},
	"b2b_outstanding_aging":      {Path: "/api/b2b/customers/outstanding", Method: "GET"},
	"b2b_gst_periods_list":       {Path: "/api/b2b/gst-periods", Method: "GET"},
	"b2b_proformas_list":         {Path: "/api/b2b/proformas", Method: "GET"},
	"b2b_proforma_get":           {Path: "/api/b2b/proformas/detail", Method: "GET"},
	"b2b_proforma_next_number":   {Path: "/api/b2b/proformas/next-number", Method: "GET"},
	"b2b_proformas_check_expiry": {Path: "/api/b2b/proformas/check-expiry", Method: "GET"},
	// Communication (WhatsApp automation)
	"whatsapp_metrics":        {Path: "/api/automation/whatsapp/metrics", Method: "GET"},
	"whatsapp_templates":      {Path: "/api/automation/whatsapp/templates", Method: "GET"},
	"whatsapp_triggers":       {Path: "/api/automation/whatsapp/triggers", Method: "GET"},
	"whatsapp_messages":       {Path: "/api/automation/whatsapp/messages", Method: "GET"},
	"whatsapp_order_messages": {Path: "/api/automation/whatsapp/messages/order", Method: "GET"},
	"whatsapp_conversations":  {Path: "/api/automation/whatsapp/conversations", Method: "GET"},
	"whatsapp_chat":           {Path: "/api/automation/whatsapp/chat", Method: "GET"},
	"whatsapp_events":         {Path: "/api/automation/whatsapp/events", Method: "GET"},
	// Marketing (Meta, SMM, Judge.me)
	"meta_overview":     {Path: "/api/marketing/meta/overview", Method: "GET"},
	"meta_campaigns":    {Path: "/api/marketing/meta/campaigns", Method: "GET"},
	"meta_adsets":       {Path: "/api/marketing/meta/adsets", Method: "GET"},
	"meta_ads":          {Path: "/api/marketing/meta/ads", Method: "GET"},
	"smm_overview":      {Path: "/api/marketing/smm/overview", Method: "GET"},
	"smm_health":        {Path: "/api/marketing/smm/health", Method: "GET"},
	"smm_post_insights": {Path: "/api/marketing/smm/post/insights", Method: "GET"},
	"smm_queue":         {Path: "/api/marketing/smm/queue", Method: "GET"},
	"smm_queue_create":  {Path: "/api/marketing/smm/queue", Method: "POST"},
	"judgeme_published": {Path: "/api/marketing/judgeme/published", Method: "GET"},
	// Feedback
	"feedback_list":          {Path: "/api/feedback", Method: "GET"},
	"feedback_config_status": {Path: "/api/feedback/config-status", Method: "GET"},
	// Abandoned checkouts
	"abandoned_checkouts_list":      {Path: "/api/abandoned-checkouts", Method: "GET"},
	"abandoned_checkouts_analytics": {Path: "/api/abandoned-checkouts/analytics", Method: "GET"},
	// Planner
	"planner_boards":    {Path: "/api/planner/boards", Method: "GET"},
	"planner_tasks":     {Path: "/api/planner/tasks", Method: "GET"},
	"planner_sprints":   {Path: "/api/planner/sprints", Method: "GET"},
	"planner_analytics": {Path: "/api/planner/analytics", Method: "GET"},
	// Support
	"support_tickets": {Path: "/api/support/tickets", Method: "GET"},
	// AI
	"ai_conversations":    {Path: "/api/ai/conversations", Method: "GET"},
	"ai_conversation_get": {Path: "/api/ai/conversations", Method: "GET"},
	// Settings
	"settings_list":       {Path: "/api/settings", Method: "GET"},
	"settings_date_range": {Path: "/api/settings/date-range", Method: "GET"},
	// System
	"system_health":    {Path: "/api/health", Method: "GET"},
	"system_docs_list": {Path: "/api/system/docs", Method: "GET"},
	"system_doc_get":   {Path: "/api/system/docs", Method: "GET"},
}

// RouteFor returns the binding for a tool name and whether it exists.
func RouteFor(name string) (RouteBinding, bool) {
	b, ok := routeMap[name]
	return b, ok
}

// Routes returns the full route map (copy).
func Routes() map[string]RouteBinding {
	out := make(map[string]RouteBinding, len(routeMap))
	for k, v := range routeMap {
		out[k] = v
	}
	return out
}

// ReadOnlyPaths returns the distinct set of internal paths served by the MCP mux.
func ReadOnlyPaths() []string {
	seen := make(map[string]struct{})
	var out []string
	for name, b := range routeMap {
		if spec, ok := DefaultCatalog.Lookup(name); ok && spec.Write {
			continue
		}
		if _, ok := seen[b.Path]; ok {
			continue
		}
		seen[b.Path] = struct{}{}
		out = append(out, b.Path)
	}
	return out
}
