package mcp

import (
	"context"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServeStdio runs the MCP server over the local stdio transport until the
// client disconnects. Used by the standalone cmd/mcp entrypoint for local dev.
func ServeStdio(ctx context.Context, server *sdk.Server) error {
	return server.Run(ctx, &sdk.StdioTransport{})
}

// HTTPHandlerOptions configures the production Streamable HTTP transport.
type HTTPHandlerOptions struct {
	// Stateless disables session persistence. Each POST creates a temporary
	// session. Recommended for this machine-key API behind a load balancer.
	Stateless bool
	// JSONResponse returns application/json rather than text/event-stream.
	JSONResponse bool
	// MaxRequestBodyBytes limits incoming request bodies.
	MaxRequestBodyBytes int64
}

// HTTPHandler returns an http.Handler exposing the MCP server over the
// Streamable HTTP transport. buildServer is invoked per new session with the
// scopes granted to the calling machine key; tools outside those scopes are
// not registered for that session, enforcing per-tool scope isolation.
func HTTPHandler(buildServer func(scopes []string) *sdk.Server, opts HTTPHandlerOptions) http.Handler {
	handlerOpts := &sdk.StreamableHTTPOptions{
		Stateless:    opts.Stateless,
		JSONResponse: opts.JSONResponse,
	}
	if opts.MaxRequestBodyBytes != 0 {
		handlerOpts.MaxRequestBodyBytes = opts.MaxRequestBodyBytes
	}
	return sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		scopes, ok := r.Context().Value("machineScopes").([]string)
		if !ok || scopes == nil {
			// HTTP machine-key sessions must never inherit the unrestricted
			// nil-scope behavior reserved for local stdio.
			scopes = []string{}
		}
		return buildServer(scopes)
	}, handlerOpts)
}
