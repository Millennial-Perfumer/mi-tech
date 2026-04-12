# Repository Guidelines

## Project Structure & Module Organization
This repository is split into three app surfaces plus infrastructure. `backend/` contains the Go API: `cmd/` for entrypoints, `internal/handler/` for HTTP transport, `internal/service/` for business logic, `internal/repository/` for persistence, and `internal/database/migrations/` for SQL migrations. `frontend/` is the main Vite + React admin UI. `frontend-mobile/` is a separate mobile-focused Vite + React client. Infra and ops files live in `nginx/`, `monitoring/`, root `docker-compose*.yml`, and [`architecture.md`](/Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/architecture.md).

## Build, Test, and Development Commands
Use the root `Makefile` for the main local workflow:

- `make install` installs frontend npm packages and backend Go modules.
- `make db-up` starts the local PostgreSQL container from `backend/docker-compose.yml`.
- `make run` starts the database, the Go API with Air reload, and the web frontend.
- `make build` builds the web frontend and compiles the backend binary.
- `cd backend && go test ./...` runs backend unit, handler, repository, and e2e tests.
- `cd frontend && npm run lint` checks the admin UI.
- `cd frontend-mobile && npm run lint` checks the mobile UI.

## Coding Style & Naming Conventions
Follow the language defaults already in the repo: Go should stay `gofmt`-formatted with package names in lowercase and tests in `*_test.go`. In React/TypeScript, keep components and screens in `PascalCase` files such as `Customers.tsx`; utility modules and APIs may use lowercase names like `api.ts`. Use 2-space indentation in frontend code and tabs/default Go formatting in backend code. ESLint is configured in both frontends; run it before opening a PR.

## Testing Guidelines
Backend tests use Go’s `testing` package with `stretchr/testify`. Prefer colocated tests next to the package they cover, and use descriptive names such as `TestWebhookService_ProcessesPaidOrder`. There is no established frontend test suite yet, so changes there should at minimum pass `npm run build` and `npm run lint`.

## Documentation Maintenance
This repository follows a strict **Doc-as-Code** mandate to ensure accuracy and reduce knowledge debt.
- **Markdown Documentation**: Every new feature or API change **MUST** be documented in the `/docs` directory.
- **Swagger / OpenAPI**: Maintain `swag` annotations in all handler files. After any change to API signatures or logic, run `swag init -g cmd/main.go --output docs/ --parseDependency --parseInternal --parseDepth 2` from the `backend/` directory.
- **Workflows**: Any update to core business logic (e.g., tax calculation or order sync) must be reflected in the respective `/docs/workflows/` files.

## Security & Configuration Tips
Do not commit secrets from `backend/.env`. Add new settings to `backend/.env.example` when needed, and prefer migration files over ad hoc database edits.

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- After modifying code files in this session, run `python3 -c "from graphify.watch import _rebuild_code; from pathlib import Path; _rebuild_code(Path('.'))"` to keep the graph current
