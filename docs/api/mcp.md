# MCP Server & Machine API Keys

A secure Model Context Protocol (MCP) server that exposes MI Tech business
capabilities to automated workflows (e.g. Codex). The catalog is read-only for
query tools and exposes explicitly scoped operational mutations for approved
machine clients. All access is
via a dedicated machine-to-machine API key — never a user JWT.

## Overview

| Item | Value |
| :--- | :--- |
| Transports | Streamable HTTP at `POST /mcp` (stateless) and stdio (`backend/cmd/mcp`) |
| Auth | `Authorization: Bearer mtk_…` machine key (API-key only) |
| Tools exposed | Read-only tools plus scoped operational write tools |
| Catalog source | `backend/internal/mcp/catalog.go` |
| Dispatch | In-process to the internal read/write muxes (`server/readonly_mux.go`) — no duplicated business logic |
| Audit | Every catalog tool invocation logged to `mcp_audit_logs` (no keys/headers stored) |
| Status | Read and scoped write catalog implemented; deployment is a separate rollout step |

## Authentication

The MCP endpoint **only** accepts machine keys (`mtk_` prefix). User JWTs and
other bearer tokens are rejected by `MachineKeyMiddleware`
(`shared/middleware/middleware.go`). On every request the middleware:

1. Rejects anything that is not an `mtk_` key.
2. Validates the SHA-256 hash, expiry, revocation, and per-minute rate limit.
3. Resolves the key's `scopes` and stores `machineKeyID` / `machineKeyName` /
   `machineScopes` on the request context.

Scopes are applied per session: only catalog tools whose scope is present in
the key's scopes are registered for that connection. Under-scoped tool calls
return an "insufficient scope" error. The `marketing:publish` scope is required
for posting to the Google Drive/n8n social queue; `marketing:read` only lists
queued posts.

## Tools & Scopes

Read-only tools map to a single `GET` route in the internal mux and carry one
read scope (e.g. `orders:read`, `customers:read`, `metrics:read`,
`gst:read`, `inventory:read`, `production:read`, `b2b:read`,
`communication:read`, `marketing:read`, `feedback:read`,
`abandoned_checkout:read`, `planner:read`, `support:read`, `ai:read`,
`settings:read`, `system:read`). Operational mutations use separate write
scopes: `orders:write`, `customers:write`, `inventory:write`,
`production:write`, `planner:write`, `b2b:write`, `communication:write`,
`marketing:write`, `feedback:write`, `support:write`, `settings:write`, and
`ai:write`. The existing social queue publisher continues to use
`marketing:publish`.

Write tools are individually allowlisted in `backend/internal/mcp/catalog.go`
and dispatched through a separate write mux to existing domain handlers. JSON
request bodies are supplied through a `payload` object; legacy endpoints that
use query identifiers expose those identifiers as explicit tool arguments.
The generic settings setter is not exposed; `settings:write` only covers the
date-range operation. Destructive tools additionally require one of
`orders:destructive`, `customers:destructive`, `inventory:destructive`,
`production:destructive`, `planner:destructive`, `b2b:destructive`,
`communication:destructive`, or `ai:destructive`. Every catalog tool
invocation is audited.

Path-based tools (e.g. `system_doc_get`) take the path segment as a named
argument and are mapped to `GET /api/.../{arg}`.

### Publishing to the Google Drive social queue

Use the `smm_queue_create` MCP tool with a caption, optional hashtags, target
platforms, and optional comma-separated public HTTPS media URLs. Supported
platforms include `instagram`, `facebook`, `threads`, and `x`. Media is
downloaded in memory, limited to 50 MB per file, and placed in the generated
Google Drive queue folder using the same flow as the web uploader. The tool
requires a machine key containing `marketing:publish`.

## Response Normalization (Phase 5)

Before a tool runs, arguments are normalized; after it returns, the payload is
sanitized:

- **Date ranges** — `start_date`/`end_date` (and camelCase variants) must be
  `YYYY-MM-DD`; a start after end is rejected with a clear error.
- **Pagination** — `limit` defaults to `50` and is capped at `500`; `page` ≥ `1`;
  `offset` ≥ `0`.
- **Sensitive-field masking** — phone, email, address, pin/zip and
  secret/token/password fields are masked (`ja••••om`) in the tool result while
  preserving JSON shape. Markdown doc responses pass through untouched.

Tool failures are returned as MCP tool errors (`CallToolResult.IsError = true`)
with the HTTP status and body; protocol-level errors (unknown tool, bad scope)
surface as transport errors.

## Connecting an LLM or Application

The MCP server speaks the **standard MCP protocol** (Streamable HTTP + stdio),
so it works with any MCP-compatible client — Claude, Codex/OpenAI, Cursor,
Windsurf, a custom script, etc. The client only needs the machine key and a
transport. Scopes are filtered per connection, so each key exposes exactly the
tools its scopes allow.

### Option A — stdio (local clients: Claude Desktop, Cursor, local scripts)

Run the bundled binary. It needs `MCP_API_KEY` (the `mtk_…` machine key) and a
database connection (`DB_DSN`, normally loaded from `backend/.env`):

```bash
# from backend/
MCP_API_KEY=mtk_xxx go run ./cmd/mcp
# or build once:
MCP_API_KEY=mtk_xxx go build -o ./bin/mcp ./cmd/mcp && ./bin/mcp
```

Claude Desktop / Cursor `mcpServers` config:

```json
{
  "mcpServers": {
    "mi-tech": {
      "command": "/abs/path/to/backend/bin/mcp",
      "env": { "MCP_API_KEY": "mtk_xxx", "DB_DSN": "postgres://postgres:password@localhost:5432/mi-tech?sslmode=disable" }
    }
  }
}
```

### Option B — Streamable HTTP (remote / cloud LLM providers)

The app already mounts the MCP endpoint at **`POST /mcp`** (stateless Streamable
HTTP, compatible with the `2025-03-26` streamable-http spec). Any MCP client that
can reach the host and send a bearer token can connect:

- **URL:** `https://<your-host>/mcp` (or `http://localhost:8080/mcp` locally)
- **Auth header:** `Authorization: Bearer mtk_xxx`
- Only `mtk_` machine keys are accepted; user JWTs are rejected.

Point your provider's MCP/tool connector at that URL with the bearer key. For
OpenAI/Codex-style HTTP MCP config:

```json
{
  "mcp_servers": {
    "mi-tech": {
      "url": "https://<your-host>/mcp",
      "headers": { "Authorization": "Bearer mtk_xxx" }
    }
  }
}
```

> [!NOTE]
> For cloud providers the server must be hosted and reachable. In production the
> MCP endpoint is served behind the existing API nginx block (TLS via certbot) at
> `https://mi-tech-api.millennialperfumer.in/mcp`. Locally, use `localhost` or the
> stdio binary.

### Creating a machine key

Keys are created by an admin (JWT) via `POST /api/mcp/keys`; the plaintext
`mtk_…` is returned once. See the key-management section below. Give each
application its own key with the minimum scopes it needs.

## Production deployment

The MCP server is part of the main backend binary (`cmd/main.go`) — no separate
service to run. The `/mcp` Streamable HTTP endpoint is mounted automatically and
protected by `MachineKeyMiddleware`. To deploy:

1. **Build & ship the backend image** (existing pipeline: `docker-compose.prod.yml`
   pulls `ghcr.io/.../mi-tech-backend:<tag>`). Migrations `131_machine_api_keys`
   and `132_mcp_audit_logs` run on startup, so the keys/audit tables appear
   automatically.
2. **nginx** already proxies the API subdomain. A dedicated `location /mcp` block
   (in `nginx/nginx.conf`) adds SSE/streaming support and CORS for browser-based
   clients. TLS is handled by certbot as for the rest of the API.
3. **Create a machine key** (below) and give the `mtk_…` value to your LLM client.

**Live endpoint:** `https://mi-tech-api.millennialperfumer.in/mcp`
**Auth:** `Authorization: Bearer mtk_…` (only `mtk_` keys accepted).

### Creating a production machine key

```bash
# 1. Admin login to obtain a JWT
JWT=$(curl -s -X POST https://mi-tech-api.millennialperfumer.in/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"'$ADMIN_USER'","password":"'$ADMIN_PASS'"}' | jq -r .token)

# 2. Create a read-only key scoped to what the client needs
curl -s -X POST https://mi-tech-api.millennialperfumer.in/api/mcp/keys \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{"name":"codex-prod","scopes":["orders:read","customers:read","metrics:read"],"rate_limit_per_min":120}' \
  | jq '.plaintext'   # the mtk_… value, shown only once
```

Then point your LLM client at the live endpoint with that key (see Option B
above, replacing `https://<your-host>/mcp` with the production URL).

## Base Path: `/api/mcp/keys`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/mcp/keys` | `GET` | 🛡️ Admin | List all machine API keys (hashes never exposed). |
| `/api/mcp/keys` | `POST` | 🛡️ Admin | Create a machine API key; plaintext returned exactly once. |
| `/api/mcp/keys/{id}` | `DELETE` | 🛡️ Admin | Revoke a machine API key. |
| `/api/mcp/keys/{id}/rotate` | `POST` | 🛡️ Admin | Replace a key's secret material; new plaintext returned once. |

> [!IMPORTANT]
> Only the SHA-256 hash of each key is stored. The plaintext `mtk_…` key is shown
> once at creation/rotation and cannot be recovered afterwards.

## 📖 Endpoint Details

### Create a Machine API Key
`POST /api/mcp/keys`

**Request Body:**
```json
{
  "name": "codex-readonly",
  "scopes": ["orders:read", "metrics:read", "gst:read"],
  "rate_limit_per_min": 60,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

**Sample Response (201):**
```json
{
  "success": true,
  "key": {
    "id": 1,
    "name": "codex-readonly",
    "scopes": ["orders:read", "metrics:read", "gst:read"],
    "rate_limit_per_min": 60,
    "expires_at": "2026-12-31T23:59:59Z",
    "revoked_at": null,
    "created_at": "2026-08-23T12:00:00Z"
  },
  "plaintext": "mtk_8v3k…",
  "key_prefix": "mtk_"
}
```

Valid scopes include the read-only MCP scopes (e.g. `orders:read`, `customers:read`,
`metrics:read`, `gst:read`, `inventory:read`, `production:read`, `b2b:read`,
`communication:read`, `marketing:read`, `feedback:read`, `abandoned_checkout:read`,
`planner:read`, `support:read`, `ai:read`, `settings:read`, `system:read`) and
the corresponding operational `:write` scopes plus the dedicated
`:destructive` scopes documented above. `marketing:publish` enables
`smm_queue_create`.

### List Machine API Keys
`GET /api/mcp/keys`

Returns all keys with metadata; `key_hash` is always omitted.

### Revoke a Machine API Key
`DELETE /api/mcp/keys/{id}`

Revoked keys are rejected immediately by `POST /mcp/*` and existing MCP sessions.

### Rotate a Machine API Key
`POST /api/mcp/keys/{id}/rotate`

Invalidates the old secret and returns a new `mtk_…` plaintext exactly once.
Metadata (name, scopes, rate limit, expiry) is preserved.

## 🔒 Security Notes

- Keys are machine-to-machine only; user JWTs are never accepted by the MCP endpoint.
- Expiry, revocation, and per-minute rate limiting are enforced on every invocation.
- Every catalog tool invocation is audited in `mcp_audit_logs` without storing keys or authorization headers.
- Sensitive fields (phone, email, address, pin/zip, secret/token/password) are masked in tool responses before they leave the server.
