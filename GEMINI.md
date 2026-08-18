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

## UI Branding & CSS Guidelines
Maintain a premium, cohesive aesthetic across the application surfaces. All UI adjustments and stylesheet edits must conform to the brand rules defined in [.agent/skills/brand-manager/SKILL.md](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/.agent/skills/brand-manager/SKILL.md):
- **Token Usage**: Never hardcode colors or spacing. Always use the CSS variables defined in [index.css](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/frontend/src/index.css) (e.g. `var(--accent-color)`, `var(--bg-color)`).
- **Motion & Transitions**: Utilize premium micro-animations and entry effects (like fadeInUp and `.hover-lift` classes) to elevate the user experience.
- **Glassmorphism**: Apply glassmorphism details for overlays and stats cards with adaptive dark theme styling.

## Testing Guidelines
Backend tests use Go’s `testing` package with `stretchr/testify`. 

> [!IMPORTANT]
> **Strict Test Directory Rule:** All backend test files (ending in `_test.go`) MUST reside only within a nested `test/` subdirectory inside the package directory they are covering (e.g., `internal/shared/middleware/test/middleware_test.go`). They must use a separate `package test` declaration and import the parent package. Under no circumstances should test files be placed in the parent package's root directory.

There is no established frontend test suite yet, so changes there should at minimum pass `npm run build` and `npm run lint`.

## Backend Concurrency & Performance
- **Queue & Batch Processing**: When iterating over a batch of items that each require external API updates or DB queries in a background queue, use `golang.org/x/sync/errgroup` to parallelize the requests.
- **Rate-Limiting Guards**: Limit concurrency to a maximum of `5` (rather than 10) to remain consistent with established rate-limiting guards and prevent overloading external API rate limits (e.g. Meta Cloud or Shopify).
- **Goroutine Safe Closures**: Always capture the loop iteration variable explicitly for goroutines inside concurrent loops (e.g. `ac := ac`).



## Documentation Maintenance
This repository follows a strict **Doc-as-Code** mandate to ensure accuracy and reduce knowledge debt.
- **Markdown Documentation**: Every new feature or API change **MUST** be documented in the `/docs` directory.
- **Swagger / OpenAPI**: Maintain `swag` annotations in all handler files. After any change to API signatures or logic, run `swag init -g cmd/main.go --output docs/ --parseDependency --parseInternal --parseDepth 2` from the `backend/` directory.
- **Workflows**: Any update to core business logic (e.g., tax calculation or order sync) must be reflected in the respective `/docs/workflows/` files.

## Security & Configuration Tips
Do not commit secrets from `backend/.env`. Add new settings to `backend/.env.example` when needed, and prefer migration files over ad hoc database edits.

