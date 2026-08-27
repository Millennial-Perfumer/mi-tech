# Meta MCP Server (Production)

The production Meta Graph API connector is exposed as Streamable HTTP at:

```
https://mcp.millennialperfumer.in/meta-mcp-server
```

It is a distinct service from the MI Tech MCP endpoint at
`https://mi-tech-api.millennialperfumer.in/mcp`.

## Required production secrets

For GitHub Actions deployments, create these as GitHub Actions **Secrets**.
The deployment workflow forwards them directly to Docker Compose on the VM; it
does not write them to the server `.env` file. For a manual deployment, export
the same values in the shell that runs `docker compose`.

```dotenv
# GitHub Secret: META_ACCESS_TOKEN
# A long-lived Meta Graph API token. Prefer a System User token for production.
META_ACCESS_TOKEN=EAAB...

# GitHub Secret: META_MCP_AUTH_TOKEN
# A random, high-entropy bearer token for clients of this MCP endpoint.
# Generate, for example: openssl rand -hex 32
META_MCP_AUTH_TOKEN=replace-with-a-random-secret

# GitHub Secret: THREADS_ACCESS_TOKEN (optional; required only for Threads).
THREADS_ACCESS_TOKEN=

# Optional repository variable. The Compose default already uses this value.
META_MCP_ALLOWED_ORIGINS=https://mcp.millennialperfumer.in
```

`META_ACCESS_TOKEN` is the token used by the server to call Meta APIs. It must
have only the permissions needed for the enabled tools. `META_MCP_AUTH_TOKEN`
is separate: it protects the public MCP endpoint and is sent by every client as
`Authorization: Bearer <META_MCP_AUTH_TOKEN>`.

The container defaults are `MCP_TRANSPORT=streamable-http`,
`MCP_HTTP_HOST=0.0.0.0`, `MCP_HTTP_PORT=3000`, and `MCP_HTTP_PATH=/mcp`; they
are set explicitly in `docker-compose.prod.yml` and do not need entries in
`.env`. The service follows the upstream `latest` image tag and Compose pulls
it again on each deployment.

## Client configuration

```json
{
  "mcpServers": {
    "meta": {
      "url": "https://mcp.millennialperfumer.in/meta-mcp-server",
      "headers": {
        "Authorization": "Bearer YOUR_META_MCP_AUTH_TOKEN"
      }
    }
  }
}
```

## Deploy and verify

1. Create a DNS record for `mcp.millennialperfumer.in` pointing at the server.
2. Add the GitHub Actions secrets above. On a fresh server, `./init-ssl.sh` now
   includes this hostname. On an existing server, expand the current certificate
   after DNS has propagated (do not rerun the bootstrap script):

   ```bash
   docker compose -f docker-compose.prod.yml run --rm certbot certonly \
     --webroot --webroot-path=/var/www/certbot --expand \
     --cert-name mi-tech.millennialperfumer.in \
     -d mi-tech.millennialperfumer.in \
     -d mi-tech-api.millennialperfumer.in \
     -d mi-tech-monitoring.millennialperfumer.in \
     -d feedback-form.millennialperfumer.in \
     -d shop-chat-agent.millennialperfumer.in \
     -d n8n.millennialperfumer.in \
     -d mcp.millennialperfumer.in
   docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
   ```
3. Deploy with the normal production compose workflow.
4. Verify the service without exposing credentials:

   ```bash
   curl -fsS https://mcp.millennialperfumer.in/healthz
   ```

The health endpoint is intentionally unauthenticated and returns only service
status. The MCP endpoint requires its bearer token. Rotate
`META_MCP_AUTH_TOKEN` whenever a client is removed; rotate the Meta token on
its own schedule or immediately after suspected exposure.

## Meta API permissions

Start with the minimum needed. Typical permissions include `pages_show_list`,
`pages_read_engagement`, `pages_manage_posts`, `instagram_basic`,
`instagram_content_publish`, `instagram_manage_insights`, `ads_read`, and
`ads_management`. Threads features additionally need a separate Threads token
and the relevant Threads permissions.
