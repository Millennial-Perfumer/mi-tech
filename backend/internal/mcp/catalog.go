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
	ScopeFeedback           = "feedback:read"
	ScopeAbandonedCheckout  = "abandoned_checkout:read"
	ScopeSupport            = "support:read"
	ScopeSettings           = "settings:read"
	ScopeSystem             = "system:read"
	ScopeOrdersWrite        = "orders:write"
	ScopeCustomersWrite     = "customers:write"
	ScopeInventoryWrite     = "inventory:write"
	ScopeProductionWrite    = "production:write"
	ScopeB2BWrite           = "b2b:write"
	ScopeCommunicationWrite = "communication:write"
	ScopeMarketingWrite     = "marketing:write"
	ScopeFeedbackWrite      = "feedback:write"
	ScopeSupportWrite       = "support:write"
	ScopeSettingsWrite      = "settings:write"
	// Destructive scopes are separate from ordinary operational writes so a
	// leaked or narrowly delegated key cannot delete/reset data by default.
	ScopeOrdersDestructive        = "orders:destructive"
	ScopeCustomersDestructive     = "customers:destructive"
	ScopeInventoryDestructive     = "inventory:destructive"
	ScopeProductionDestructive    = "production:destructive"
	ScopeB2BDestructive           = "b2b:destructive"
	ScopeCommunicationDestructive = "communication:destructive"
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
		Name:        "customers_get",
		Description: "Fetch one complete current customer profile, including contact, address, source, and lifetime order metrics.",
		Scope:       ScopeCustomers,
		Route:       "/api/customers/",
		PathArgs:    []string{"id"},
		Args: []ArgSpec{
			argReq("id", ArgInt, "Internal customer id."),
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

	// --- Amazon ---
	{
		Name:        "amazon_listings",
		Description: "Fetch a live page of Amazon listings using the configured Seller ID and marketplace ID. Returns the exact matching count and a next page token when more results are available.",
		Scope:       ScopeInventory,
		Route:       "/api