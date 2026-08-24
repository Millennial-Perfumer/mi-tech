package mcp

// ArgType enumerates the JSON-schema types an MCP tool argument may have.
type ArgType string

const (
	ArgString ArgType = "string"
	ArgInt    ArgType = "integer"
	ArgNumber ArgType = "number"
)

// ArgSpec describes a single tool input argument. It drives both the MCP JSON
// input schema and the internal GET query-parameter mapping.
type ArgSpec struct {
	Name        string
	Type        ArgType
	Required    bool
	Description string
	Default     any
}

// ToolSpec is the single source of truth for one MCP tool. Route is the
// internal backend path that the tool maps to; Write identifies the explicit
// publishing exception to the read-only catalog.
type ToolSpec struct {
	Name        string
	Description string
	Scope       string
	Route       string
	Args        []ArgSpec
	// PathArgs lists argument names that are injected into the URL path
	// (appended after Route) instead of the query string. Order matters.
	PathArgs []string
	// Write marks the small set of explicitly authorized MCP mutations.
	Write bool
}

// Catalog is an ordered collection of tool specs.
type Catalog []ToolSpec

// Scope constants. All MCP scopes are read-only in this release.
const (
	ScopeOrders            = "orders:read"
	ScopeCustomers         = "customers:read"
	ScopeMetrics           = "metrics:read"
	ScopeGST               = "gst:read"
	ScopeInventory         = "inventory:read"
	ScopeProduction        = "production:read"
	ScopeB2B               = "b2b:read"
	ScopeCommunication     = "communication:read"
	ScopeMarketing         = "marketing:read"
	ScopeMarketingPublish  = "marketing:publish"
	ScopeFeedback          = "feedback:read"
	ScopeAbandonedCheckout = "abandoned_checkout:read"
	ScopePlanner           = "planner:read"
	ScopeSupport           = "support:read"
	ScopeAI                = "ai:read"
	ScopeSettings          = "settings:read"
	ScopeSystem            = "system:read"
)

// arg is a shorthand constructor for an optional ArgSpec.
func arg(name string, typ ArgType, desc string) ArgSpec {
	return ArgSpec{Name: name, Type: typ, Description: desc}
}

// argReq is a shorthand constructor for a required ArgSpec.
func argReq(name string, typ ArgType, desc string) ArgSpec {
	return ArgSpec{Name: name, Type: typ, Required: true, Description: desc}
}

// DefaultCatalog defines every read-only tool exposed by the MCP server.
// Each entry maps 1:1 to a read-only backend GET route (see route_map.go).
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
