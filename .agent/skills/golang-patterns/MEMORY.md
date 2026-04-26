# MI-Tech Go Backend Memory

## Architecture
- **Entry point**: `backend/cmd/main.go`
- **Hot reload**: Air (`go run github.com/air-verse/air@latest -c .air.toml`)
- **Layered architecture**: `handler/ → service/ → repository/` with `entity/` for domain models
- **Database**: PostgreSQL via GORM, migrations in `internal/database/migrations/`
- **Config**: `internal/config/config.go` loads `.env`, `SettingsProvider` reads from `app_configs` DB table with ENV fallback

## Key Patterns
- **Repository pattern**: All repos accept `*gorm.DB`, most use interfaces (e.g., `repository.OrderRepository`)
- **SettingsProvider**: Dynamic config from DB with `os.Getenv()` fallback. Used by Shopify, Amazon, WhatsApp clients
- **SyncOrchestrator**: Central inventory coordinator — adjusts stock across Shopify + Amazon when orders change
- **Delta-based inventory**: `OrderRepository.syncInventoryDeltas()` compares old vs new line items on upsert
- **Webhook-driven**: Shopify pushes order events; Amazon uses background polling (`AmazonOrderPoller`)

## External Clients
- `internal/client/shopify/` — GraphQL Admin API, uses `SettingsProvider`
- `internal/client/amazon/` — SP-API with STS/SigV4 signing, uses `SettingsProvider`
- WhatsApp automation in `internal/automation/whatsapp/`

## Build Constraints
- Go binary is at `/usr/local/go/bin/go`
- Must use sandboxed cache: `GOMODCACHE=$(pwd)/.gocache/mod GOCACHE=$(pwd)/.gocache/build`
- Must set `CGO_ENABLED=0` and `GOFLAGS=-buildvcs=false`
- Multiple `main` packages live in `cmd/` subdirectories (e.g., `cmd/seed/`, `cmd/verify_db/`)

## Testing
- Uses `stretchr/testify`, colocated `*_test.go` files
- Mocks in `internal/service/mocks/`
- Run: `go test ./...` with the cache env vars above
