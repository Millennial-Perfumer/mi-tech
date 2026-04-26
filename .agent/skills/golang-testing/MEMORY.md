# MI-Tech Go Testing Memory

## Setup
- **Framework**: Go `testing` package + `stretchr/testify` (assert, mock)
- **Mocks**: `internal/service/mocks/` directory
- **Run command**: 
  ```bash
  export GOMODCACHE=$(pwd)/.gocache/mod
  export GOCACHE=$(pwd)/.gocache/build
  export GOFLAGS=-buildvcs=false
  export CGO_ENABLED=0
  /usr/local/go/bin/go test ./...
  ```

## Conventions
- Tests colocated with source: `webhook_service.go` → `webhook_service_test.go`
- Naming: `TestWebhookService_ProcessesPaidOrder`, `TestOrderRepository_Upsert`
- Table-driven tests for handler validation
- Repository tests use real DB (integration style) or mock interfaces

## Key Test Areas
- **Handler tests**: HTTP request/response validation
- **Service tests**: Business logic with mocked repositories
- **Repository tests**: GORM queries against test database
- **Webhook tests**: Signature verification + event processing

## Test Utilities
- `internal/testutil/` — helper functions for test setup
- `entity.StrPtr()` — convenience for pointer-to-string in test fixtures
