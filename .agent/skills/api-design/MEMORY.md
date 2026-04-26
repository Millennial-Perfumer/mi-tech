# MI-Tech API Design Memory

## Architecture
- **Transport**: `net/http` ServeMux (no framework, no Gin/Echo)
- **Routing**: Centralized in `internal/server/routes.go`
- **Auth**: JWT Bearer tokens, middleware in `internal/handler/auth_middleware.go`
- **Roles**: `admin`, `read` — parsed from JWT payload
- **API prefix**: `/api/`

## Key Endpoints
| Prefix | Handler | Purpose |
|--------|---------|---------|
| `/api/orders` | OrderHandler | CRUD + search, pagination |
| `/api/shopify/sync` | SyncHandler | Trigger Shopify order sync |
| `/api/amazon/sync` | SyncHandler | Trigger Amazon order sync |
| `/api/configs` | ConfigsHandler | App configuration CRUD |
| `/api/settings` | SettingsHandler | User preferences |
| `/api/inventory` | InventoryHandler | SKU management, stock levels |
| `/api/customers` | CustomerHandler | Customer CRUD |
| `/api/webhooks` | WebhookHandler | Shopify webhook receiver |
| `/api/automation` | AutomationHandler | WhatsApp templates & messages |
| `/api/reports` | ReportHandler | GST reports |
| `/api/metrics` | MetricsHandler | Dashboard stats |
| `/api/feedback` | FeedbackHandler | Customer sentiment |

## Patterns
- Standard JSON response: `{"success": true, "data": {...}}` or `{"success": false, "error": "..."}`
- Pagination: `?page=1&limit=25&sort_by=created_at&sort_order=DESC`
- Search: `?search=query`
- Filters: `?source=shopify&financial_status=paid&fulfillment_status=unfulfilled`
- Swagger/OpenAPI: Maintained via `swag` annotations, output in `docs/`

## Integration APIs
- **Shopify**: GraphQL Admin API (orders, products, inventory levels)
- **Amazon**: SP-API REST (orders, listings, reports) with STS SigV4 signing
- **WhatsApp**: Meta Cloud API (message templates, webhook handling)
- **Meta Marketing**: Graph API (ad accounts, campaigns)
