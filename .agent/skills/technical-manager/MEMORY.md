# MI-Tech Technical Manager Memory

## Project Overview
**MI-Tech** is a multi-channel e-commerce operations platform for Millennial Perfumer — an Indian D2C perfume brand. It manages orders, inventory, customers, invoicing, and automation across Shopify, Amazon India, and WhatsApp.

## Architecture at a Glance
```
┌─────────────┐  ┌──────────────┐  ┌───────────────────┐
│  frontend/  │  │frontend-mob/ │  │frontend-feedback/ │
│ Vite+React  │  │ Vite+React   │  │  Vite+React       │
│ Admin UI    │  │ Mobile Client│  │  Feedback Forms    │
└──────┬──────┘  └──────┬───────┘  └───────┬───────────┘
       │                │                   │
       └────────────────┼───────────────────┘
                        │ REST/JSON
                ┌───────┴────────┐
                │   backend/     │
                │   Go API       │
                │   (net/http)   │
                └───────┬────────┘
                        │
          ┌─────────────┼─────────────┐
          │             │             │
    ┌─────┴─────┐ ┌─────┴────┐ ┌─────┴─────┐
    │ PostgreSQL│ │ Shopify  │ │  Amazon   │
    │   (GORM)  │ │ GraphQL  │ │  SP-API   │
    └───────────┘ └──────────┘ └───────────┘
```

## Key Business Flows

### 1. Order Lifecycle
```
Shopify Webhook → WebhookHandler → OrderService.UpsertOrder()
                                          ↓
                                   OrderRepository.Upsert()
                                          ↓
                                   syncInventoryDeltas() ← Delta-based (new - old)
                                          ↓
                                   SyncOrchestrator.AdjustStock()
                                          ↓
                              ┌───────────┴───────────┐
                              │                       │
                        Shopify Update          Amazon Update
                    (GraphQL mutation)      (Listings PATCH API)
```

### 2. Amazon Polling
```
AmazonOrderPoller (every 3 min)
    → GetOrders(LastUpdatedAfter: -6h)
    → For each order: GetOrderItems()
    → processDeduction() or processReversal()
    → orderRepo.Upsert()
```

### 3. Configuration Flow
```
Frontend SettingsTab → PUT /api/configs/:key
    → ConfigsRepository.Set()
    → SettingsProvider.Get() (reads from app_configs table)
    → Client uses SettingsProvider dynamically
```

## Dependency Graph (What Touches What)
| Change Area | Affects |
|-------------|---------|
| `entity/` models | repositories, services, handlers, DTOs |
| `repository/` | services that depend on it |
| `service/` | handlers, orchestrator |
| `config/settings_provider.go` | ALL clients (Shopify, Amazon, WhatsApp, Meta) |
| `server.go` | Wiring — touch when adding new service/handler |
| `App.tsx` tabs | Only the specific tab section |
| `SettingsTab.tsx` | Only settings display |
| `app_configs` migration | SettingsProvider + SettingsTab category |
| `api_design` routes | Handler + frontend API calls |

## Common Task Patterns

### Adding a New Feature (Full Stack)
1. `database-migrations` — Schema/config changes
2. `golang-patterns` — Repository → Service → Handler
3. `api-design` — Endpoint design
4. `frontend-patterns` — React component + API call
5. `frontend-design` — Styling and UX
6. `golang-testing` — Tests
7. `systematic-debugging` — Build verification

### Adding a New Integration
1. `api-connector-builder` — Follow existing pattern
2. `database-migrations` — Config keys in app_configs
3. `golang-patterns` — SettingsProvider methods
4. `frontend-patterns` — SettingsTab category
5. `security-review` — Credential handling

### Fixing a Bug
1. `systematic-debugging` — Reproduce and diagnose
2. `golang-patterns` or `frontend-patterns` — Fix
3. `golang-testing` — Regression test
4. `verification-before-completion` — Verify

### UI Redesign
1. `brainstorming` — Design exploration
2. `frontend-design` — Visual implementation
3. `accessibility` — WCAG compliance
4. `browser-qa` — Visual verification

## Lessons Learned
- Always use sandboxed Go build (GOMODCACHE, GOCACHE, CGO_ENABLED=0)
- `verify_pinnacle.go` in `cmd/` caused redeclared main — keep utility commands in subdirectories
- Amazon client was refactored from static `*config.Config` to dynamic `*config.SettingsProvider`
- Inventory sync uses delta-based logic: compares old line items vs new before adjusting stock
- The frontend `make frontend` terminal is usually already running — don't start a new one
- Migration files are never deleted, only appended. All must be idempotent (ON CONFLICT DO NOTHING)
- WhatsApp OTP delivery depends on `automation_messages` table with correct foreign keys
