# MI Tech Platform

MI Tech is a multi-service commerce operations platform. The repository contains the primary Go API, two React frontends, a Shopify chat-agent application, and production observability infrastructure.

## Services

| Service | Path | Technology | Responsibility |
| --- | --- | --- | --- |
| API | `backend/` | Go, PostgreSQL | Core business domains, integrations, authentication, invoices, and webhooks. |
| Operations UI | `frontend/` | React, Vite | Main internal operations interface. |
| Feedback UI | `frontend-feedback/` | React, Vite | Customer feedback experience. |
| Shopify chat agent | `shop-chat-agent/` | React Router, Prisma | Shopify product-listing assistant. |
| Review checker | `amazon-review-checker/` | Go | Standalone Amazon-review utility. |
| Observability | `monitoring/` | Prometheus, Grafana, Loki, Tempo | Metrics, logs, and tracing configuration. |

## Local development

Prerequisites: Go 1.25, Node.js 20 for the React frontends, Node.js 18 for `shop-chat-agent`, Docker Compose, and npm.

```sh
cp backend/.env.example backend/.env
make install
make db-up
make run
```

Use `make build` to build the API and React frontends. The Shopify chat agent is independent: `cd shop-chat-agent && npm ci && npm run dev`.

## Delivery model

Pull requests run backend validation, React lint/build checks, Shopify app lint/typecheck/build checks, and CodeQL analysis. Manual deployments build every production service, publish commit-SHA images, and deploy those immutable image references. Production secrets, including database credentials and `IMAGE_TAG`, are required environment variables; they are never supplied with defaults.

## Documentation

Start with [the technical documentation](docs/index.md). Architecture decisions belong in `docs/adr/`; API and workflow documentation must be updated with business-logic or API changes.
