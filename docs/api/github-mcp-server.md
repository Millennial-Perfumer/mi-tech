# GitHub MCP Server (Production)

The published GitHub connector image is deployed as a separate Streamable HTTP
service behind the existing MCP Nginx host:

```text
https://mcp.millennialperfumer.in/github-mcp-server
```

The Meta connector remains at
`https://mcp.millennialperfumer.in/meta-mcp-server`.

## Required runtime values

Set these as GitHub Actions organization secrets for the normal deployment
workflow:

```dotenv
# Organization secret: PAT_GITHUB
# Fine-grained GitHub PAT with only the repository/org permissions required.
PAT_GITHUB=github_pat_...

# Organization secret shared by the Meta and GitHub MCP services.
MCP_HTTP_AUTH_TOKEN=replace-with-a-random-secret
```

The workflow maps the PAT to the container environment and passes the shared
HTTP token directly to Compose:

```dotenv
GITHUB_PERSONAL_ACCESS_TOKEN=github_pat_...
MCP_HTTP_AUTH_TOKEN=replace-with-a-random-secret
```

`GITHUB_PERSONAL_ACCESS_TOKEN` authorizes calls to GitHub. Clients must not send
that PAT. They authenticate to the remote MCP endpoint with the shared token:

```http
Authorization: Bearer <MCP_HTTP_AUTH_TOKEN>
```

## Container settings

The production Compose service sets these non-secret values:

```dotenv
MCP_TRANSPORT=streamable-http
MCP_HTTP_HOST=0.0.0.0
GITHUB_MCP_HTTP_PORT=3001
GITHUB_MCP_HTTP_PATH=/github/mcp
MCP_HTTP_ALLOWED_ORIGINS=https://mcp.millennialperfumer.in
```

The image also supports optional GitHub Enterprise configuration:

```dotenv
GITHUB_HOST=github.example.com
GITHUB_API_URL=https://github.example.com/api/v3
GITHUB_UPLOADS_URL=https://uploads.github.com
```

Only add those values when using GitHub Enterprise Server. The image’s bundled
upstream binary and native API tools use the same PAT and do not require a
Docker socket.

## Verification

The health endpoint is intentionally unauthenticated:

```bash
curl -fsS https://mcp.millennialperfumer.in/github-mcp-healthz
```

The MCP endpoint requires the separate bearer token and a normal Streamable
HTTP MCP initialize handshake:

```json
{
  "mcpServers": {
    "github": {
      "url": "https://mcp.millennialperfumer.in/github-mcp-server",
      "headers": {
        "Authorization": "Bearer YOUR_MCP_HTTP_AUTH_TOKEN"
      }
    }
  }
}
```

The public package image is `ghcr.io/open-work-org/github-mcp-server:latest`.
For reproducible releases, replace `latest` in
`docker-compose.prod.yml` with a published release tag or immutable digest.
