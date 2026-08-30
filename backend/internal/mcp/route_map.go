package mcp

// RouteBinding describes how an MCP tool maps to an internal backend route.
type RouteBinding struct {
	// Path is the internal route (no query string).
	Path string
	// Method is the HTTP method used by the internal MCP mux.
	Method string
}

// routeMap is the single source of truth for the allowlisted route surface.
// It is keyed by tool name and validated against DefaultCatalog in tests.
var routeMap = map[string]RouteBinding{
	// Orders
	"orders_list":    {Path: "/api/orders", Method: "GET"},
	"orders_get":     {Path: "/api/orders", Method: "GET"},
	"orders_history": {Path: "/api/orders/history", Method: "GET"},
	"orders_sources": {Path: "/api/sources", Method: "GET"},
	// Customers
	"customers_list":    {Path: "/api/customers", Method: "GET"},
	"customers_history": {Path: "/api/customers/history", Method: "GET"},
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
	"inventory_logs":      {Path: "/api/inventory/logs", Method: "GET"},
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
	// Orders and customers (write)
	"orders_create":                {Path: "/api/orders", Method: "POST"},
	"orders_update":                {Path: "/api/orders", Method: "PUT"},
	"orders_update_status":         {Path: "/api/orders/status", Method: "PUT"},
	"orders_update_payment_status": {Path: "/api/orders/payment-status", Method: "PUT"},
	"orders_mark_delivered":        {Path: "/api/orders/delivered", Method: "PUT"},
	"customers_create":             {Path: "/api/customers", Method: "POST"},
	"customers_bulk_delete":        {Path: "/api/customers/bulk-delete", Method: "POST"},
	"customers_update":             {Path: "/api/customers/", Method: "PUT"},
	"customers_delete":             {Path: "/api/customers/", Method: "DELETE"},
	// Product inventory (write)
	"inventory_create":         {Path: "/api/inventory", Method: "POST"},
	"inventory_bulk_create":    {Path: "/api/inventory/bulk", Method: "POST"},
	"inventory_update_item":    {Path: "/api/inventory/item", Method: "PUT"},
	"inventory_set_stock":      {Path: "/api/inventory/stock", Method: "POST"},
	"inventory_adjust_stock":   {Path: "/api/inventory/adjust", Method: "POST"},
	"inventory_create_mapping": {Path: "/api/inventory/map", Method: "POST"},
	"inventory_delete_mapping": {Path: "/api/inventory/map", Method: "DELETE"},
	"inventory_sync_shopify":   {Path: "/api/inventory/sync-shopify", Method: "POST"},
	"inventory_sync_prices":    {Path: "/api/inventory/sync-prices", Method: "POST"},
	"inventory_sync_amazon":    {Path: "/api/inventory/amazon/sync", Method: "POST"},
	"inventory_clear":          {Path: "/api/inventory", Method: "DELETE"},
	// Production (write)
	"oils_create":                 {Path: "/api/inventory/oil", Method: "POST"},
	"oils_update":                 {Path: "/api/inventory/oil", Method: "PUT"},
	"oils_delete":                 {Path: "/api/inventory/oil", Method: "DELETE"},
	"oils_bulk_delete":            {Path: "/api/inventory/oil/bulk-delete", Method: "POST"},
	"suppliers_create":            {Path: "/api/inventory/suppliers", Method: "POST"},
	"suppliers_update":            {Path: "/api/inventory/suppliers", Method: "PUT"},
	"suppliers_delete":            {Path: "/api/inventory/suppliers", Method: "DELETE"},
	"purchase_orders_create":      {Path: "/api/inventory/po", Method: "POST"},
	"purchase_orders_bulk_create": {Path: "/api/inventory/po/bulk", Method: "POST"},
	"purchase_orders_update":      {Path: "/api/inventory/po", Method: "PUT"},
	"purchase_orders_delete":      {Path: "/api/inventory/po", Method: "DELETE"},
	"manufacturing_create":        {Path: "/api/inventory/manufacturing", Method: "POST"},
	"manufacturing_update":        {Path: "/api/inventory/manufacturing", Method: "PUT"},
	"manufacturing_delete":        {Path: "/api/inventory/manufacturing", Method: "DELETE"},
	// Planner (write)
	"planner_task_create":   {Path: "/api/planner/tasks", Method: "POST"},
	"planner_task_update":   {Path: "/api/planner/tasks", Method: "PUT"},
	"planner_task_delete":   {Path: "/api/planner/tasks", Method: "DELETE"},
	"planner_task_move":     {Path: "/api/planner/tasks/move", Method: "POST"},
	"planner_sprint_create": {Path: "/api/planner/sprints", Method: "POST"},
	"planner_sprint_update": {Path: "/api/planner/sprints", Method: "PUT"},
	"planner_sprint_delete": {Path: "/api/planner/sprints", Method: "DELETE"},
	// Synchronization and settings (write)
	"shopify_sync_orders":     {Path: "/api/shopify/sync", Method: "POST"},
	"shopify_reset_orders":    {Path: "/api/shopify/reset", Method: "POST"},
	"settings_set_date_range": {Path: "/api/settings/date-range", Method: "PUT"},
	// Support and abandoned checkouts (write)
	"support_ticket_create":            {Path: "/api/support/tickets", Method: "POST"},
	"support_ticket_update":            {Path: "/api/support/tickets/", Method: "PUT"},
	"abandoned_checkout_recover":       {Path: "/api/abandoned-checkouts/recover", Method: "POST"},
	"abandoned_checkout_update_status": {Path: "/api/abandoned-checkouts/status", Method: "PUT"},
	"abandoned_checkout_delete":        {Path: "/api/abandoned-checkouts", Method: "DELETE"},
	// B2B (write)
	"b2b_customer_create":                {Path: "/api/b2b/customers", Method: "POST"},
	"b2b_customer_update":                {Path: "/api/b2b/customers", Method: "PUT"},
	"b2b_customer_delete":                {Path: "/api/b2b/customers", Method: "DELETE"},
	"b2b_invoice_create":                 {Path: "/api/b2b/invoices", Method: "POST"},
	"b2b_invoice_update":                 {Path: "/api/b2b/invoices", Method: "PUT"},
	"b2b_invoice_delete":                 {Path: "/api/b2b/invoices", Method: "DELETE"},
	"b2b_invoice_issue":                  {Path: "/api/b2b/invoices/issue", Method: "POST"},
	"b2b_invoice_cancel":                 {Path: "/api/b2b/invoices/cancel", Method: "POST"},
	"b2b_invoice_deduct_inventory":       {Path: "/api/b2b/invoices/deduct-inventory", Method: "POST"},
	"b2b_invoice_revert_inventory":       {Path: "/api/b2b/invoices/revert-inventory", Method: "POST"},
	"b2b_invoice_update_payment":         {Path: "/api/b2b/invoices/payment", Method: "POST"},
	"b2b_payment_terms_create_or_update": {Path: "/api/b2b/payment-terms", Method: "POST"},
	"b2b_payment_terms_update":           {Path: "/api/b2b/payment-terms", Method: "PUT"},
	"b2b_credit_note_create":             {Path: "/api/b2b/credit-notes", Method: "POST"},
	"b2b_credit_note_update":             {Path: "/api/b2b/credit-notes", Method: "PUT"},
	"b2b_credit_note_delete":             {Path: "/api/b2b/credit-notes", Method: "DELETE"},
	"b2b_credit_note_issue":              {Path: "/api/b2b/credit-notes/issue", Method: "POST"},
	"b2b_credit_note_cancel":             {Path: "/api/b2b/credit-notes/cancel", Method: "POST"},
	"b2b_debit_note_create":              {Path: "/api/b2b/debit-notes", Method: "POST"},
	"b2b_debit_note_update":              {Path: "/api/b2b/debit-notes", Method: "PUT"},
	"b2b_debit_note_delete":              {Path: "/api/b2b/debit-notes", Method: "DELETE"},
	"b2b_debit_note_issue":               {Path: "/api/b2b/debit-notes/issue", Method: "POST"},
	"b2b_debit_note_cancel":              {Path: "/api/b2b/debit-notes/cancel", Method: "POST"},
	"b2b_proforma_create":                {Path: "/api/b2b/proformas", Method: "POST"},
	"b2b_proforma_update":                {Path: "/api/b2b/proformas", Method: "PUT"},
	"b2b_proforma_delete":                {Path: "/api/b2b/proformas", Method: "DELETE"},
	"b2b_proforma_issue":                 {Path: "/api/b2b/proformas/issue", Method: "POST"},
	"b2b_proforma_accept":                {Path: "/api/b2b/proformas/accept", Method: "POST"},
	"b2b_proforma_reject":                {Path: "/api/b2b/proformas/reject", Method: "POST"},
	"b2b_proforma_cancel":                {Path: "/api/b2b/proformas/cancel", Method: "POST"},
	"b2b_proforma_create_revision":       {Path: "/api/b2b/proformas/revision", Method: "POST"},
	"b2b_proforma_convert_to_invoice":    {Path: "/api/b2b/proformas/convert", Method: "POST"},
	"b2b_proformas_mark_expired":         {Path: "/api/b2b/proformas/check-expiry", Method: "POST"},
	// WhatsApp automation (write)
	"whatsapp_template_create":          {Path: "/api/automation/whatsapp/templates", Method: "POST"},
	"whatsapp_template_update":          {Path: "/api/automation/whatsapp/templates", Method: "PUT"},
	"whatsapp_template_delete":          {Path: "/api/automation/whatsapp/templates", Method: "DELETE"},
	"whatsapp_templates_sync_status":    {Path: "/api/automation/whatsapp/templates/sync", Method: "POST"},
	"whatsapp_templates_sync_all":       {Path: "/api/automation/whatsapp/templates/sync-all", Method: "POST"},
	"whatsapp_template_sync_single":     {Path: "/api/automation/whatsapp/templates/sync-single", Method: "POST"},
	"whatsapp_template_fetch":           {Path: "/api/automation/whatsapp/templates/fetch", Method: "POST"},
	"whatsapp_trigger_create":           {Path: "/api/automation/whatsapp/triggers", Method: "POST"},
	"whatsapp_trigger_update":           {Path: "/api/automation/whatsapp/triggers", Method: "PUT"},
	"whatsapp_trigger_delete":           {Path: "/api/automation/whatsapp/triggers", Method: "DELETE"},
	"whatsapp_send_message":             {Path: "/api/automation/whatsapp/send-message", Method: "POST"},
	"whatsapp_send_manual":              {Path: "/api/automation/whatsapp/send-manual", Method: "POST"},
	"whatsapp_send_bulk":                {Path: "/api/automation/whatsapp/send-bulk", Method: "POST"},
	"whatsapp_conversation_mode_update": {Path: "/api/automation/whatsapp/conversations/mode", Method: "PUT"},
	"whatsapp_event_create":             {Path: "/api/automation/whatsapp/events", Method: "POST"},
	"whatsapp_event_delete":             {Path: "/api/automation/whatsapp/events", Method: "DELETE"},
	"whatsapp_metrics_sync":             {Path: "/api/automation/whatsapp/sync-metrics", Method: "POST"},
	// Social marketing and reviews (write)
	"smm_post":                 {Path: "/api/marketing/smm/post", Method: "POST"},
	"smm_sync":                 {Path: "/api/marketing/smm/sync", Method: "POST"},
	"judgeme_generate_reviews": {Path: "/api/marketing/judgeme/generate", Method: "POST"},
	"judgeme_submit_reviews":   {Path: "/api/marketing/judgeme/submit", Method: "POST"},
	// Feedback and AI (write)
	"feedback_bulk_send":             {Path: "/api/feedback/bulk-send", Method: "POST"},
	"feedback_update_comment":        {Path: "/api/orders/feedback/comment", Method: "PUT"},
	"feedback_post_judgeme":          {Path: "/api/orders/feedback/post-judgeme", Method: "POST"},
	"feedback_request_google_review": {Path: "/api/orders/feedback/request-google-review", Method: "POST"},
	"ai_chat":                        {Path: "/api/ai/chat", Method: "POST"},
	"ai_conversation_delete":         {Path: "/api/ai/conversations", Method: "DELETE"},
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
