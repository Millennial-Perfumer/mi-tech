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

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit style, especially `feat:` commits. Keep using prefixes like `feat:`, `fix:`, and `chore:` with a concise subject. Pull requests should explain the user-visible change, note any migration or config impact, link the relevant issue, and include screenshots for frontend changes. If you touch schema or automation flows, call that out explicitly in the PR description.

## Security & Configuration Tips
Do not commit secrets from `backend/.env`. Add new settings to `backend/.env.example` when needed, and prefer migration files over ad hoc database edits.
