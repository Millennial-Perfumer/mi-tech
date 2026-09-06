package mcp

// ArgType enumerates the JSON-schema types an MCP tool argument may have.
type ArgType string

const (
	ArgString  ArgType = "string"
	ArgInt     ArgType = "integer"
	ArgNumber  ArgType = "number"
	ArgObject  ArgType = "object"
	ArgBoolean ArgType = "boolean"
	ArgArray   ArgType = "array"
)

// ArgSpec describes a single tool input argument. It drives the MCP JSON input
// schema and the internal request mapping.
type ArgSpec struct {
	Name        string
	Type        ArgType
	Required    bool
	Description string
	Default     any
}

// ToolSpec is the single source of truth for one MCP tool. Route and Method
// identify the internal backend operation; Write identifies a mutating tool.
type ToolSpec struct {
	Name        string
	Description string
	Scope       string
	Route       string
	Args        []ArgSpec
	// PathArgs lists argument names that are injected into the URL path
	// (appended after Route) instead of the query string. Order matters.
	PathArgs []string
	// QueryArgs lists arguments sent as query parameters for write tools.
	QueryArgs []string
	// Method is the HTTP method used for dispatch. Empty means GET.
	Method string
	// Write marks an explicitly authorized MCP mutation.
	Write bool
}

// Catalog is an ordered collection of tool specs.
type Catalog []ToolSpec

// Scope constants. Read and write scopes are intentionally separate so a
// machine key can be granted reporting access without mutation access.
const (
	ScopeOrders             = "orders:read"
	ScopeCustomers          = "customers:read"
	ScopeMetrics            = "metrics:read"
	ScopeGST                = "gst:read"
	ScopeInventory          = "inventory:read"
	ScopeProduction         = "production:read"
	ScopeB2B                = "b2b:read"
	ScopeCommunication      = "communication:read"
	ScopeMarketing          = "marketing:read"
	ScopeMarketingPublish   = "marketing:publish"
	ScopeFeedback           = "feedback:read"
	ScopeAbandonedCheckout  = "abandoned_checkout:read"
	ScopePlanner            = "planner:read"
	ScopeSupport            = "support:read"
	ScopeAI                 = "ai:read"
	ScopeSettings           = "settings:read"
	ScopeSystem             = "system:read"
	ScopeOrdersWrite        = "orders:write"
	ScopeCustomersWrite     = "customers:write"
	ScopeInventoryWrite     = "inventory:write"
	ScopeProductionWrite    = "production:write"
	ScopePlannerWrite       = "planner:write"
	ScopeB2BWrite           = "b2b:write"
	ScopeCommunicationWrite = "communication:write"
	ScopeMarketingWrite     = "marketing:write"
	ScopeFeedbackWrite      = "feedback:write"
	ScopeSupportWrite       = "support:write"
	ScopeSettingsWrite      = "settings:write"
	ScopeAIWrite            = "ai:write"
	// Destructive scopes are separate from ordinary operational writes so a
	// leaked or narrowly delegated key cannot delete/reset data by default.
	ScopeOrdersDestructive        = "orders:destructive"
	ScopeCustomersDestructive     = "customers:destructive"
	ScopeInventoryDestructive     = "inventory:destructive"
	ScopeProductionDestructive    = "production:destructive"
	ScopePlannerDestructive       = "planner:destructive"
	ScopeB2BDestructive           = "b2b:destructive"
	ScopeCommunicationDestructive = "communication:destructive"
	ScopeAIDestructive            = "ai:destructive"
)

// arg is a shorthand constructor for an optional ArgSpec.
func arg(name string, typ ArgType, desc string) ArgSpec {
	return ArgSpec{Name: name, Type: typ, Description: desc}
}

// argReq is a shorthand constructor for a required ArgSpec.
func argReq(name string, typ ArgType, desc string) ArgSpec {
	return ArgSpec{Name: name, Type: typ, Required: true, Description: desc}
}

// DefaultCatalog defines every tool exposed by the MCP server. Each entry maps
// 1:1 to an allowlisted backend route (see route_map.go).
var DefaultCatalog = Catalog{
	// --- Orders ---
	{
		Name:        "orders_list",
		Description: "List orders with optional date range, pagination, search, and status filters.",
		Scope:       ScopeOrders,
		Route:       "/api/orders",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date (YYYY-MM-DD) start of the range."),
			arg("end_date", ArgString, "ISO date (YYYY-MM-DD) end of the range."),
			arg("page", ArgInt, "Page number, 1-based."),
			arg("limit", ArgInt, "Number of results per page."),
			arg("search", ArgString, "Free-text search across order fields."),
			arg("source", ArgString, "Order source filter."),
			arg("financial_status", ArgString, "Financial status filter."),
			arg("fulfillment_status", ArgString, "Fulfillment status filter."),
			arg("status", ArgString, "Order status filter."),
			arg("sort_by", ArgString, "Field to sort by."),
			arg("sort_order", ArgString, "Sort direction (asc/desc)."),
			arg("state", ArgString, "Order state filter."),
		},
	},
	{
		Name:        "orders_get",
		Description: "Fetch a single order by id.",
		Scope:       ScopeOrders,
		Route:       "/api/orders",
		Args: []ArgSpec{
			argReq("id", ArgInt, "Order id."),
		},
	},
	{
		Name:        "orders_history",
		Description: "List immutable order history, including previous AWBs/tracking values, status changes, customer-detail changes, and sync events.",
		Scope:       ScopeOrders,
		Route:       "/api/orders/history",
		Args: []ArgSpec{
			arg("id", ArgInt, "Internal order id."),
			arg("external_order_id", ArgString, "External order id."),
			arg("search", ArgString, "Search current and historical event values, including old AWBs."),
			arg("event_type", ArgString, "Filter by event type."),
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("page", ArgInt, "Page number, 1-based."),
			arg("limit", ArgInt, "Number of events per page, maximum 100."),
		},
	},
	{
		Name:        "orders_sources",
		Description: "List distinct order sources.",
		Scope:       ScopeOrders,
		Route:       "/api/sources",
	},

	// --- Customers ---
	{
		Name:        "customers_list",
		Description: "List customers with pagination, search, and spend/order filters.",
		Scope:       ScopeCustomers,
		Route:       "/api/customers",
		Args: []ArgSpec{
			arg("page", ArgInt, "Page number."),
			arg("pageSize", ArgInt, "Page size."),
			arg("search", ArgString, "Free-text search."),
			arg("sortBy", ArgString, "Field to sort by."),
			arg("sortOrder", ArgString, "Sort direction (asc/desc)."),
			arg("source_id", ArgString, "Source id filter."),
			arg("min_spent", ArgNumber, "Minimum total spent."),
			arg("max_spent", ArgNumber, "Maximum total spent."),
			arg("min_orders", ArgInt, "Minimum order count."),
			arg("city", ArgString, "City filter."),
			arg("state", ArgString, "State filter."),
		},
	},
	{
		Name:        "customers_history",
		Description: "List immutable customer profile history, including previous contact and address values.",
		Scope:       ScopeCustomers,
		Route:       "/api/customers/history",
		Args: []ArgSpec{
			arg("id", ArgInt, "Internal customer id."),
			arg("order_id", ArgInt, "Order id that caused the customer change."),
			arg("phone", ArgString, "Customer phone number."),
			arg("search", ArgString, "Search historical customer values."),
			arg("event_type", ArgString, "Filter by event type."),
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("page", ArgInt, "Page number, 1-based."),
			arg("limit", ArgInt, "Number of events per page, maximum 100."),
		},
	},

	// --- Dashboard metrics ---
	{
		Name:        "dashboard_metrics",
		Description: "Aggregate dashboard metrics (revenue, orders, customers) for a date range.",
		Scope:       ScopeMetrics,
		Route:       "/api/dashboard/metrics",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
		},
	},
	{
		Name:        "dashboard_top_products",
		Description: "Top products by revenue for a date range.",
		Scope:       ScopeMetrics,
		Route:       "/api/dashboard/top-products",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
			arg("limit", ArgInt, "Number of products to return (default 5)."),
		},
	},
	{
		Name:        "dashboard_revenue_trend",
		Description: "Daily revenue trend for a date range.",
		Scope:       ScopeMetrics,
		Route:       "/api/dashboard/revenue-trend",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
		},
	},
	{
		Name:        "dashboard_geo_distribution",
		Description: "Customer/order distribution by geography.",
		Scope:       ScopeMetrics,
		Route:       "/api/dashboard/geo-distribution",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
			arg("limit", ArgInt, "Number of regions to return (default 5)."),
		},
	},

	// --- GST reports ---
	{
		Name:        "gst_summary",
		Description: "GST summary report for a date range.",
		Scope:       ScopeGST,
		Route:       "/api/reports/summary",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
		},
	},
	{
		Name:        "gst_state_wise",
		Description: "GST summary broken down by state.",
		Scope:       ScopeGST,
		Route:       "/api/reports/state-wise",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
		},
	},
	{
		Name:        "gst_hsn_wise",
		Description: "GST summary broken down by HSN code.",
		Scope:       ScopeGST,
		Route:       "/api/reports/hsn-wise",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
		},
	},
	{
		Name:        "gst_documents_issued",
		Description: "Count of GST documents issued by type and date range.",
		Scope:       ScopeGST,
		Route:       "/api/reports/documents-issued",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("source_ids", ArgString, "Comma-separated source ids."),
		},
	},
	{
		Name:        "gst_gstr1_json",
		Description: "GSTR-1 JSON export for a date range, optionally filtered by GSTIN.",
		Scope:       ScopeGST,
		Route:       "/api/reports/gstr1-json",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
			arg("gstin", ArgString, "Optional GSTIN filter."),
		},
	},

	// --- Inventory ---
	{
		Name:        "inventory_dashboard",
		Description: "Inventory dashboard page with search, sort, and pagination.",
		Scope:       ScopeInventory,
		Route:       "/api/inventory",
		Args: []ArgSpec{
			arg("search", ArgString, "Free-text search."),
			arg("page", ArgInt, "Page number."),
			arg("limit", ArgInt, "Page size."),
			arg("sort", ArgString, "Sort expression."),
		},
	},
	{
		Name:        "inventory_logs",
		Description: "List stock movement history by inventory item id or external order id, including stock before and after when recorded.",
		Scope:       ScopeInventory,
		Route:       "/api/inventory/logs",
		Args: []ArgSpec{
			arg("id", ArgInt, "Inventory item id."),
			arg("external_order_id", ArgString, "External order id that caused the movement."),
		},
	},
	{
		Name:        "inventory_next_sku",
		Description: "Generate the next available inventory SKU.",
		Scope:       ScopeInventory,
		Route:       "/api/inventory/next-sku",
	},

	// --- Production: suppliers, oils, purchase orders, manufacturing ---
	{
		Name:        "suppliers_list",
		Description: "List production suppliers.",
		Scope:       ScopeProduction,
		Route:       "/api/inventory/suppliers",
	},
	{
		Name:        "oils_list",
		Description: "List oil inventory items.",
		Scope:       ScopeProduction,
		Route:       "/api/inventory/oil",
	},
	{
		Name:        "purchase_orders_list",
		Description: "List purchase orders; supports recent-days and recent-limit views.",
		Scope:       ScopeProduction,
		Route:       "/api/inventory/po",
		Args: []ArgSpec{
			arg("days", ArgInt, "Return purchase orders grouped by the last N days."),
			arg("page", ArgInt, "Page number for the recent-days view."),
			arg("limit", ArgInt, "Return the most recent N purchase orders."),
		},
	},
	{
		Name:        "manufacturing_list",
		Description: "List manufacturing batches.",
		Scope:       ScopeProduction,
		Route:       "/api/inventory/manufacturing",
	},

	// --- B2B billing ---
	{
		Name:        "b2b_customers_list",
		Description: "List B2B customers with optional search.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/customers",
		Args: []ArgSpec{
			arg("search", ArgString, "Free-text search."),
		},
	},
	{
		Name:        "b2b_invoices_list",
		Description: "List B2B invoices with optional date range and status filter.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/invoices",
		Args: []ArgSpec{
			arg("startDate", ArgString, "ISO date start."),
			arg("endDate", ArgString, "ISO date end."),
			arg("status", ArgString, "Invoice status filter."),
		},
	},
	{
		Name:        "b2b_invoice_get",
		Description: "Fetch a single B2B invoice by id.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/invoices/detail",
		Args: []ArgSpec{
			argReq("id", ArgInt, "Invoice id."),
		},
	},
	{
		Name:        "b2b_invoice_next_number",
		Description: "Compute the next B2B invoice number.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/invoices/next-number",
	},
	{
		Name:        "b2b_payment_terms_list",
		Description: "List B2B payment terms.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/payment-terms",
	},
	{
		Name:        "b2b_credit_notes_list",
		Description: "List B2B credit notes, optionally filtered by invoice.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/credit-notes",
		Args: []ArgSpec{
			arg("invoice_id", ArgInt, "Optional invoice id filter."),
		},
	},
	{
		Name:        "b2b_debit_notes_list",
		Description: "List B2B debit notes, optionally filtered by invoice.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/debit-notes",
		Args: []ArgSpec{
			arg("invoice_id", ArgInt, "Optional invoice id filter."),
		},
	},
	{
		Name:        "b2b_customer_ledger",
		Description: "Fetch a B2B customer's ledger.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/customers/ledger",
		Args: []ArgSpec{
			argReq("customer_id", ArgInt, "Customer id."),
		},
	},
	{
		Name:        "b2b_outstanding_aging",
		Description: "Outstanding aging report for B2B customers.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/customers/outstanding",
	},
	{
		Name:        "b2b_gst_periods_list",
		Description: "List B2B GST periods.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/gst-periods",
	},
	{
		Name:        "b2b_proformas_list",
		Description: "List B2B proforma invoices with optional date range and status filter.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/proformas",
		Args: []ArgSpec{
			arg("startDate", ArgString, "ISO date start."),
			arg("endDate", ArgString, "ISO date end."),
			arg("status", ArgString, "Proforma status filter."),
		},
	},
	{
		Name:        "b2b_proforma_get",
		Description: "Fetch a single B2B proforma by id.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/proformas/detail",
		Args: []ArgSpec{
			argReq("id", ArgInt, "Proforma id."),
		},
	},
	{
		Name:        "b2b_proforma_next_number",
		Description: "Compute the next B2B proforma number.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/proformas/next-number",
	},
	{
		Name:        "b2b_proformas_check_expiry",
		Description: "Check and report expired B2B proformas.",
		Scope:       ScopeB2B,
		Route:       "/api/b2b/proformas/check-expiry",
	},

	// --- Communication (WhatsApp automation) ---
	{
		Name:        "whatsapp_metrics",
		Description: "WhatsApp automation metrics for a date range.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/metrics",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "whatsapp_templates",
		Description: "List WhatsApp message templates for a date range.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/templates",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "whatsapp_triggers",
		Description: "List WhatsApp automation triggers.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/triggers",
	},
	{
		Name:        "whatsapp_messages",
		Description: "List WhatsApp automation messages.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/messages",
	},
	{
		Name:        "whatsapp_order_messages",
		Description: "List WhatsApp messages for orders.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/messages/order",
	},
	{
		Name:        "whatsapp_conversations",
		Description: "List WhatsApp conversations.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/conversations",
	},
	{
		Name:        "whatsapp_chat",
		Description: "List messages in a WhatsApp conversation.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/chat",
		Args: []ArgSpec{
			argReq("conversation_id", ArgInt, "Conversation id."),
			arg("limit", ArgInt, "Number of messages (default 50)."),
			arg("offset", ArgInt, "Offset for pagination."),
		},
	},
	{
		Name:        "whatsapp_events",
		Description: "List WhatsApp automation events.",
		Scope:       ScopeCommunication,
		Route:       "/api/automation/whatsapp/events",
	},

	// --- Marketing (Meta, SMM, Judge.me) ---
	{
		Name:        "meta_overview",
		Description: "Meta marketing overview for a date range.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/meta/overview",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "meta_campaigns",
		Description: "Meta marketing campaigns for a date range.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/meta/campaigns",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "meta_adsets",
		Description: "Meta marketing ad sets for a date range.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/meta/adsets",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "meta_ads",
		Description: "Meta marketing ads for a date range.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/meta/ads",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "smm_overview",
		Description: "Social media management overview, optionally per platform.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/smm/overview",
		Args: []ArgSpec{
			arg("platform", ArgString, "Platform filter."),
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "smm_health",
		Description: "Social media integration health check.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/smm/health",
	},
	{
		Name:        "smm_post_insights",
		Description: "Insights for a specific social post.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/smm/post/insights",
		Args: []ArgSpec{
			argReq("id", ArgString, "Post id."),
			arg("media_type", ArgString, "Media type filter."),
		},
	},
	{
		Name:        "smm_queue",
		Description: "List queued social media posts.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/smm/queue",
	},
	{
		Name:        "smm_queue_create",
		Description: "Queue a social post for Google Drive/n8n publishing. Accepts a caption, hashtags, target platforms, and optional public HTTPS media URLs.",
		Scope:       ScopeMarketingPublish,
		Route:       "/api/marketing/smm/queue",
		Write:       true,
		Args: []ArgSpec{
			argReq("caption", ArgString, "Post caption."),
			arg("hashtags", ArgString, "Hashtags to include in the post."),
			arg("post_type", ArgString, "SINGLE_PHOTO, CAROUSEL, or VIDEO. Inferred when omitted."),
			arg("target_platforms", ArgString, "Comma-separated platforms: instagram, facebook, threads, x."),
			arg("media_urls", ArgString, "Comma-separated public HTTPS media URLs."),
		},
	},
	{
		Name:        "judgeme_published",
		Description: "List published Judge.me reviews.",
		Scope:       ScopeMarketing,
		Route:       "/api/marketing/judgeme/published",
		Args: []ArgSpec{
			arg("page", ArgInt, "Page number."),
			arg("limit", ArgInt, "Page size."),
			arg("product_id", ArgString, "Product id filter."),
			arg("search", ArgString, "Free-text search."),
		},
	},

	// --- Feedback ---
	{
		Name:        "feedback_list",
		Description: "List customer feedback, optionally filtered by order or phone.",
		Scope:       ScopeFeedback,
		Route:       "/api/feedback",
		Args: []ArgSpec{
			arg("o", ArgString, "Order id filter."),
			arg("p", ArgString, "Phone number filter."),
		},
	},
	{
		Name:        "feedback_config_status",
		Description: "Feedback configuration status.",
		Scope:       ScopeFeedback,
		Route:       "/api/feedback/config-status",
	},

	// --- Abandoned checkouts ---
	{
		Name:        "abandoned_checkouts_list",
		Description: "List abandoned checkouts with filters and pagination.",
		Scope:       ScopeAbandonedCheckout,
		Route:       "/api/abandoned-checkouts",
		Args: []ArgSpec{
			arg("page", ArgInt, "Page number."),
			arg("limit", ArgInt, "Page size."),
			arg("search", ArgString, "Free-text search."),
			arg("status", ArgString, "Status filter."),
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},
	{
		Name:        "abandoned_checkouts_analytics",
		Description: "Abandoned checkout analytics for a date range.",
		Scope:       ScopeAbandonedCheckout,
		Route:       "/api/abandoned-checkouts/analytics",
		Args: []ArgSpec{
			arg("start_date", ArgString, "ISO date start."),
			arg("end_date", ArgString, "ISO date end."),
		},
	},

	// --- Planner ---
	{
		Name:        "planner_boards",
		Description: "List planner boards.",
		Scope:       ScopePlanner,
		Route:       "/api/planner/boards",
	},
	{
		Name:        "planner_tasks",
		Description: "List planner tasks with optional filters.",
		Scope:       ScopePlanner,
		Route:       "/api/planner/tasks",
		Args: []ArgSpec{
			arg("board_id", ArgInt, "Board id filter."),
			arg("sprint_id", ArgString, "Sprint id filter."),
			arg("status", ArgString, "Status filter."),
			arg("priority", ArgString, "Priority filter."),
			arg("search", ArgString, "Free-text search."),
		},
	},
	{
		Name:        "planner_sprints",
		Description: "List planner sprints, optionally by status.",
		Scope:       ScopePlanner,
		Route:       "/api/planner/sprints",
		Args: []ArgSpec{
			arg("status", ArgString, "Status filter."),
		},
	},
	{
		Name:        "planner_analytics",
		Description: "Planner analytics for a sprint and/or task.",
		Scope:       ScopePlanner,
		Route:       "/api/planner/analytics",
		Args: []ArgSpec{
			arg("sprint_id", ArgInt, "Sprint id."),
			arg("task_id", ArgInt, "Task id."),
		},
	},

	// --- Support ---
	{
		Name:        "support_tickets",
		Description: "List support tickets.",
		Scope:       ScopeSupport,
		Route:       "/api/support/tickets",
	},

	// --- AI ---
	{
		Name:        "ai_conversations",
		Description: "List AI conversation history.",
		Scope:       ScopeAI,
		Route:       "/api/ai/conversations",
	},
	{
		Name:        "ai_conversation_get",
		Description: "Fetch a single AI conversation by id.",
		Scope:       ScopeAI,
		Route:       "/api/ai/conversations",
		Args: []ArgSpec{
			argReq("id", ArgInt, "Conversation id."),
		},
	},

	// --- Settings (safe, masked) ---
	{
		Name:        "settings_list",
		Description: "List application settings. Secret values are masked.",
		Scope:       ScopeSettings,
		Route:       "/api/settings",
	},
	{
		Name:        "settings_date_range",
		Description: "Get the current configured date range.",
		Scope:       ScopeSettings,
		Route:       "/api/settings/date-range",
	},

	// --- System health & docs ---
	{
		Name:        "system_health",
		Description: "System health check.",
		Scope:       ScopeSystem,
		Route:       "/api/health",
	},
	{
		Name:        "system_docs_list",
		Description: "List available documentation slugs.",
		Scope:       ScopeSystem,
		Route:       "/api/system/docs",
	},
	{
		Name:        "system_doc_get",
		Description: "Fetch documentation content by slug.",
		Scope:       ScopeSystem,
		Route:       "/api/system/docs",
		Args: []ArgSpec{
			argReq("slug", ArgString, "Document slug."),
		},
		PathArgs: []string{"slug"},
	},
}

// writeTool creates an explicitly allowlisted mutation. The payload is sent
// as the JSON request body; IDs used by legacy DELETE/status endpoints remain
// query arguments so existing handlers can be reused unchanged.
func writeTool(name, description, scope, method, route string, queryArgs ...string) ToolSpec {
	args := []ArgSpec{argReq("payload", ArgObject, "JSON request body accepted by the MI-Tech API handler.")}
	for _, q := range queryArgs {
		args = append(args, argReq(q, ArgString, "Identifier or query value."))
	}
	return ToolSpec{Name: name, Description: description, Scope: scope, Method: method, Route: route, QueryArgs: queryArgs, Args: args, Write: true}
}

func writeToolOptionalPayload(name, description, scope, method, route string, queryArgs ...string) ToolSpec {
	args := []ArgSpec{arg("payload", ArgObject, "Optional JSON request body accepted by the MI-Tech API handler.")}
	for _, q := range queryArgs {
		args = append(args, argReq(q, ArgString, "Identifier or query value."))
	}
	return ToolSpec{Name: name, Description: description, Scope: scope, Method: method, Route: route, QueryArgs: queryArgs, Args: args, Write: true}
}

func writeToolPath(name, description, scope, method, route, pathArg string) ToolSpec {
	return ToolSpec{Name: name, Description: description, Scope: scope, Method: method, Route: route, PathArgs: []string{pathArg}, Args: []ArgSpec{argReq("payload", ArgObject, "JSON request body accepted by the MI-Tech API handler."), argReq(pathArg, ArgString, "Resource identifier.")}, Write: true}
}

func writeToolPathNoBody(name, description, scope, method, route, pathArg string) ToolSpec {
	return ToolSpec{Name: name, Description: description, Scope: scope, Method: method, Route: route, PathArgs: []string{pathArg}, Args: []ArgSpec{argReq(pathArg, ArgString, "Resource identifier.")}, Write: true}
}

func writeToolNoBody(name, description, scope, method, route string, queryArgs ...string) ToolSpec {
	args := make([]ArgSpec, 0, len(queryArgs))
	for _, q := range queryArgs {
		args = append(args, argReq(q, ArgString, "Identifier or query value."))
	}
	return ToolSpec{Name: name, Description: description, Scope: scope, Method: method, Route: route, QueryArgs: queryArgs, Args: args, Write: true}
}

func init() {
	DefaultCatalog = append(DefaultCatalog,
		// Orders and customers
		writeTool("orders_create", "Create an order.", ScopeOrdersWrite, "POST", "/api/orders"),
		writeTool("orders_update", "Update an order.", ScopeOrdersWrite, "PUT", "/api/orders", "id"),
		writeTool("orders_update_status", "Update an order status.", ScopeOrdersWrite, "PUT", "/api/orders/status", "id"),
		writeTool("orders_update_payment_status", "Update an order payment status.", ScopeOrdersWrite, "PUT", "/api/orders/payment-status", "id"),
		writeToolNoBody("orders_mark_delivered", "Mark an order as delivered.", ScopeOrdersWrite, "PUT", "/api/orders/delivered", "id"),
		writeTool("customers_create", "Create a customer.", ScopeCustomersWrite, "POST", "/api/customers"),
		writeTool("customers_bulk_delete", "Delete customers in bulk.", ScopeCustomersDestructive, "POST", "/api/customers/bulk-delete"),
		writeToolPath("customers_update", "Update a customer.", ScopeCustomersWrite, "PUT", "/api/customers/", "id"),
		writeToolPathNoBody("customers_delete", "Delete a customer by id.", ScopeCustomersDestructive, "DELETE", "/api/customers/", "id"),
		// Product inventory and stock synchronization
		writeTool("inventory_create", "Create an inventory item.", ScopeInventoryWrite, "POST", "/api/inventory"),
		writeTool("inventory_bulk_create", "Create inventory items in bulk.", ScopeInventoryWrite, "POST", "/api/inventory/bulk"),
		writeTool("inventory_update_item", "Update an inventory item.", ScopeInventoryWrite, "PUT", "/api/inventory/item"),
		writeToolNoBody("inventory_set_stock", "Set stock to an exact quantity.", ScopeInventoryWrite, "POST", "/api/inventory/stock", "id", "val"),
		writeToolNoBody("inventory_adjust_stock", "Adjust stock by a delta.", ScopeInventoryWrite, "POST", "/api/inventory/adjust", "id", "delta"),
		writeTool("inventory_create_mapping", "Create an external inventory mapping.", ScopeInventoryWrite, "POST", "/api/inventory/map"),
		writeToolNoBody("inventory_delete_mapping", "Delete an external inventory mapping.", ScopeInventoryDestructive, "DELETE", "/api/inventory/map", "id"),
		writeTool("inventory_sync_shopify", "Synchronize inventory with Shopify.", ScopeInventoryWrite, "POST", "/api/inventory/sync-shopify"),
		writeTool("inventory_sync_prices", "Synchronize inventory prices.", ScopeInventoryWrite, "POST", "/api/inventory/sync-prices"),
		writeToolOptionalPayload("inventory_sync_amazon", "Synchronize inventory with Amazon, optionally for a date range.", ScopeInventoryWrite, "POST", "/api/inventory/amazon/sync"),
		writeToolNoBody("inventory_clear", "Clear all inventory. Use only for an intentional warehouse reset.", ScopeInventoryDestructive, "DELETE", "/api/inventory"),
		// Production: oils, suppliers, purchase orders, manufacturing
		writeTool("oils_create", "Create an oil inventory item.", ScopeProductionWrite, "POST", "/api/inventory/oil"),
		writeTool("oils_update", "Update an oil inventory item.", ScopeProductionWrite, "PUT", "/api/inventory/oil"),
		writeToolNoBody("oils_delete", "Delete an oil inventory item.", ScopeProductionDestructive, "DELETE", "/api/inventory/oil", "id"),
		writeTool("oils_bulk_delete", "Delete oil inventory items in bulk.", ScopeProductionDestructive, "POST", "/api/inventory/oil/bulk-delete"),
		writeTool("suppliers_create", "Create a supplier.", ScopeProductionWrite, "POST", "/api/inventory/suppliers"),
		writeTool("suppliers_update", "Update a supplier.", ScopeProductionWrite, "PUT", "/api/inventory/suppliers"),
		writeToolNoBody("suppliers_delete", "Delete a supplier.", ScopeProductionDestructive, "DELETE", "/api/inventory/suppliers", "id"),
		writeTool("purchase_orders_create", "Create a purchase order.", ScopeProductionWrite, "POST", "/api/inventory/po"),
		writeTool("purchase_orders_bulk_create", "Create purchase orders in bulk.", ScopeProductionWrite, "POST", "/api/inventory/po/bulk"),
		writeTool("purchase_orders_update", "Update a purchase order.", ScopeProductionWrite, "PUT", "/api/inventory/po"),
		writeToolNoBody("purchase_orders_delete", "Delete a purchase order.", ScopeProductionDestructive, "DELETE", "/api/inventory/po", "id"),
		writeTool("manufacturing_create", "Create a manufacturing record.", ScopeProductionWrite, "POST", "/api/inventory/manufacturing"),
		writeTool("manufacturing_update", "Update a manufacturing record.", ScopeProductionWrite, "PUT", "/api/inventory/manufacturing"),
		writeToolNoBody("manufacturing_delete", "Delete a manufacturing record.", ScopeProductionDestructive, "DELETE", "/api/inventory/manufacturing", "id"),
		// Planner
		writeTool("planner_task_create", "Create a planner task.", ScopePlannerWrite, "POST", "/api/planner/tasks"),
		writeTool("planner_task_update", "Update a planner task.", ScopePlannerWrite, "PUT", "/api/planner/tasks", "id"),
		writeToolNoBody("planner_task_delete", "Delete a planner task.", ScopePlannerDestructive, "DELETE", "/api/planner/tasks", "id"),
		writeTool("planner_task_move", "Move a planner task.", ScopePlannerWrite, "POST", "/api/planner/tasks/move"),
		writeTool("planner_sprint_create", "Create a planner sprint.", ScopePlannerWrite, "POST", "/api/planner/sprints"),
		writeTool("planner_sprint_update", "Update a planner sprint.", ScopePlannerWrite, "PUT", "/api/planner/sprints", "id"),
		writeToolNoBody("planner_sprint_delete", "Delete a planner sprint.", ScopePlannerDestructive, "DELETE", "/api/planner/sprints", "id"),
		// Synchronization and configuration
		writeToolOptionalPayload("shopify_sync_orders", "Synchronize orders from Shopify, optionally for a date range.", ScopeOrdersWrite, "POST", "/api/shopify/sync"),
		writeToolNoBody("shopify_reset_orders", "Reset synchronized orders. Destructive operation.", ScopeOrdersDestructive, "POST", "/api/shopify/reset"),
		writeTool("settings_set_date_range", "Set the application date range.", ScopeSettingsWrite, "PUT", "/api/settings/date-range"),
		// Support and abandoned checkout operations
		writeTool("support_ticket_create", "Create a support ticket.", ScopeSupportWrite, "POST", "/api/support/tickets"),
		writeToolPath("support_ticket_update", "Update a support ticket status.", ScopeSupportWrite, "PUT", "/api/support/tickets/", "id"),
		writeTool("abandoned_checkout_recover", "Recover an abandoned checkout.", ScopeOrdersWrite, "POST", "/api/abandoned-checkouts/recover"),
		writeTool("abandoned_checkout_update_status", "Update an abandoned checkout status.", ScopeOrdersWrite, "PUT", "/api/abandoned-checkouts/status"),
		writeToolNoBody("abandoned_checkout_delete", "Delete an abandoned checkout.", ScopeOrdersDestructive, "DELETE", "/api/abandoned-checkouts", "id"),
		// B2B customers, invoices, notes, and proformas
		writeTool("b2b_customer_create", "Create a B2B customer.", ScopeB2BWrite, "POST", "/api/b2b/customers"),
		writeTool("b2b_customer_update", "Update a B2B customer.", ScopeB2BWrite, "PUT", "/api/b2b/customers"),
		writeToolNoBody("b2b_customer_delete", "Delete a B2B customer.", ScopeB2BDestructive, "DELETE", "/api/b2b/customers", "id"),
		writeTool("b2b_invoice_create", "Create a B2B invoice.", ScopeB2BWrite, "POST", "/api/b2b/invoices"),
		writeTool("b2b_invoice_update", "Update a B2B invoice.", ScopeB2BWrite, "PUT", "/api/b2b/invoices"),
		writeToolNoBody("b2b_invoice_delete", "Delete a B2B invoice.", ScopeB2BDestructive, "DELETE", "/api/b2b/invoices", "id"),
		writeToolNoBody("b2b_invoice_issue", "Issue a B2B invoice.", ScopeB2BWrite, "POST", "/api/b2b/invoices/issue", "id"),
		writeToolNoBody("b2b_invoice_cancel", "Cancel a B2B invoice.", ScopeB2BDestructive, "POST", "/api/b2b/invoices/cancel", "id"),
		writeToolNoBody("b2b_invoice_deduct_inventory", "Deduct inventory for a B2B invoice.", ScopeB2BWrite, "POST", "/api/b2b/invoices/deduct-inventory", "id"),
		writeToolNoBody("b2b_invoice_revert_inventory", "Revert inventory deduction for a B2B invoice.", ScopeB2BWrite, "POST", "/api/b2b/invoices/revert-inventory", "id"),
		writeTool("b2b_invoice_update_payment", "Record a B2B invoice payment.", ScopeB2BWrite, "POST", "/api/b2b/invoices/payment"),
		writeTool("b2b_payment_terms_create_or_update", "Create or update B2B payment terms.", ScopeB2BWrite, "POST", "/api/b2b/payment-terms"),
		writeTool("b2b_payment_terms_update", "Update B2B payment terms.", ScopeB2BWrite, "PUT", "/api/b2b/payment-terms"),
		writeTool("b2b_credit_note_create", "Create a B2B credit note.", ScopeB2BWrite, "POST", "/api/b2b/credit-notes"),
		writeTool("b2b_credit_note_update", "Update a B2B credit note.", ScopeB2BWrite, "PUT", "/api/b2b/credit-notes"),
		writeToolNoBody("b2b_credit_note_delete", "Delete a B2B credit note.", ScopeB2BDestructive, "DELETE", "/api/b2b/credit-notes", "id"),
		writeToolNoBody("b2b_credit_note_issue", "Issue a B2B credit note.", ScopeB2BWrite, "POST", "/api/b2b/credit-notes/issue", "id"),
		writeToolNoBody("b2b_credit_note_cancel", "Cancel a B2B credit note.", ScopeB2BDestructive, "POST", "/api/b2b/credit-notes/cancel", "id"),
		writeTool("b2b_debit_note_create", "Create a B2B debit note.", ScopeB2BWrite, "POST", "/api/b2b/debit-notes"),
		writeTool("b2b_debit_note_update", "Update a B2B debit note.", ScopeB2BWrite, "PUT", "/api/b2b/debit-notes"),
		writeToolNoBody("b2b_debit_note_delete", "Delete a B2B debit note.", ScopeB2BDestructive, "DELETE", "/api/b2b/debit-notes", "id"),
		writeToolNoBody("b2b_debit_note_issue", "Issue a B2B debit note.", ScopeB2BWrite, "POST", "/api/b2b/debit-notes/issue", "id"),
		writeToolNoBody("b2b_debit_note_cancel", "Cancel a B2B debit note.", ScopeB2BDestructive, "POST", "/api/b2b/debit-notes/cancel", "id"),
		writeTool("b2b_proforma_create", "Create a B2B proforma invoice.", ScopeB2BWrite, "POST", "/api/b2b/proformas"),
		writeTool("b2b_proforma_update", "Update a B2B proforma invoice.", ScopeB2BWrite, "PUT", "/api/b2b/proformas"),
		writeToolNoBody("b2b_proforma_delete", "Delete a B2B proforma invoice.", ScopeB2BDestructive, "DELETE", "/api/b2b/proformas", "id"),
		writeToolNoBody("b2b_proforma_issue", "Issue a B2B proforma invoice.", ScopeB2BWrite, "POST", "/api/b2b/proformas/issue", "id"),
		writeToolNoBody("b2b_proforma_accept", "Accept a B2B proforma invoice.", ScopeB2BWrite, "POST", "/api/b2b/proformas/accept", "id"),
		writeToolNoBody("b2b_proforma_reject", "Reject a B2B proforma invoice.", ScopeB2BWrite, "POST", "/api/b2b/proformas/reject", "id"),
		writeToolNoBody("b2b_proforma_cancel", "Cancel a B2B proforma invoice.", ScopeB2BDestructive, "POST", "/api/b2b/proformas/cancel", "id"),
		writeToolNoBody("b2b_proforma_create_revision", "Create a proforma invoice revision.", ScopeB2BWrite, "POST", "/api/b2b/proformas/revision", "id"),
		writeToolNoBody("b2b_proforma_convert_to_invoice", "Convert a proforma invoice to a tax invoice.", ScopeB2BWrite, "POST", "/api/b2b/proformas/convert", "id"),
		writeToolNoBody("b2b_proformas_mark_expired", "Mark expired proforma invoices.", ScopeB2BWrite, "POST", "/api/b2b/proformas/check-expiry"),
		// WhatsApp automation
		writeTool("whatsapp_template_create", "Create a WhatsApp template.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/templates"),
		writeTool("whatsapp_template_update", "Update a WhatsApp template mapping.", ScopeCommunicationWrite, "PUT", "/api/automation/whatsapp/templates"),
		writeToolNoBody("whatsapp_template_delete", "Delete a WhatsApp template.", ScopeCommunicationDestructive, "DELETE", "/api/automation/whatsapp/templates", "id"),
		writeToolNoBody("whatsapp_templates_sync_status", "Synchronize WhatsApp template statuses from Meta.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/templates/sync"),
		writeToolNoBody("whatsapp_templates_sync_all", "Synchronize all WhatsApp templates.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/templates/sync-all"),
		writeToolNoBody("whatsapp_template_sync_single", "Synchronize one WhatsApp template.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/templates/sync-single", "name"),
		writeToolNoBody("whatsapp_template_fetch", "Fetch a WhatsApp template from Meta.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/templates/fetch", "name"),
		writeTool("whatsapp_trigger_create", "Create a WhatsApp automation trigger.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/triggers"),
		writeTool("whatsapp_trigger_update", "Enable or disable a WhatsApp automation trigger.", ScopeCommunicationWrite, "PUT", "/api/automation/whatsapp/triggers"),
		writeToolNoBody("whatsapp_trigger_delete", "Delete a WhatsApp automation trigger.", ScopeCommunicationDestructive, "DELETE", "/api/automation/whatsapp/triggers", "id"),
		writeTool("whatsapp_send_message", "Send a free-text WhatsApp message.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/send-message"),
		writeTool("whatsapp_send_manual", "Send a manual WhatsApp message.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/send-manual"),
		writeTool("whatsapp_send_bulk", "Send a bulk WhatsApp marketing message.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/send-bulk"),
		writeTool("whatsapp_conversation_mode_update", "Change a WhatsApp conversation between auto and human mode.", ScopeCommunicationWrite, "PUT", "/api/automation/whatsapp/conversations/mode"),
		writeTool("whatsapp_event_create", "Create a WhatsApp automation event.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/events"),
		writeToolNoBody("whatsapp_event_delete", "Delete a WhatsApp automation event.", ScopeCommunicationDestructive, "DELETE", "/api/automation/whatsapp/events", "id"),
		writeToolNoBody("whatsapp_metrics_sync", "Synchronize WhatsApp automation metrics.", ScopeCommunicationWrite, "POST", "/api/automation/whatsapp/sync-metrics"),
		// Social marketing and reviews
		writeTool("smm_post", "Publish content to a social platform.", ScopeMarketingWrite, "POST", "/api/marketing/smm/post"),
		writeToolNoBody("smm_sync", "Synchronize social marketing metrics and history.", ScopeMarketingWrite, "POST", "/api/marketing/smm/sync", "platform"),
		writeTool("judgeme_generate_reviews", "Generate Judge.me review drafts.", ScopeMarketingWrite, "POST", "/api/marketing/judgeme/generate"),
		writeTool("judgeme_submit_reviews", "Submit Judge.me reviews.", ScopeMarketingWrite, "POST", "/api/marketing/judgeme/submit"),
		// Feedback and AI
		writeTool("feedback_bulk_send", "Send feedback requests for selected orders.", ScopeFeedbackWrite, "POST", "/api/feedback/bulk-send"),
		writeTool("feedback_update_comment", "Update an internal feedback comment.", ScopeFeedbackWrite, "PUT", "/api/orders/feedback/comment", "id"),
		writeToolNoBody("feedback_post_judgeme", "Post a feedback review to Judge.me.", ScopeFeedbackWrite, "POST", "/api/orders/feedback/post-judgeme", "id"),
		writeToolNoBody("feedback_request_google_review", "Request a Google review from feedback.", ScopeFeedbackWrite, "POST", "/api/orders/feedback/request-google-review", "id"),
		writeTool("ai_chat", "Send a message to the MI-Tech AI assistant.", ScopeAIWrite, "POST", "/api/ai/chat"),
		writeToolNoBody("ai_conversation_delete", "Delete an AI conversation.", ScopeAIDestructive, "DELETE", "/api/ai/conversations", "id"),
	)
}

// Lookup returns the tool spec with the given name and whether it was found.
func (c Catalog) Lookup(name string) (ToolSpec, bool) {
	for _, spec := range c {
		if spec.Name == name {
			return spec, true
		}
	}
	return ToolSpec{}, false
}

// Scopes returns the distinct set of scopes required by the catalog.
func (c Catalog) Scopes() []string {
	seen := make(map[string]struct{})
	var out []string
	for _, spec := range c {
		if _, ok := seen[spec.Scope]; ok {
			continue
		}
		seen[spec.Scope] = struct{}{}
		out = append(out, spec.Scope)
	}
	return out
}
