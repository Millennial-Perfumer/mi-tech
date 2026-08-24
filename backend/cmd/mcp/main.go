// MCP server entrypoint for the read-only MI Tech MCP server.
//
// Local stdio mode (default):
//
//	MCP_API_KEY=mtk_... go run ./cmd/mcp
//
// The key must be a machine API key created via the admin API
// (POST /api/mcp/keys). Its scopes determine which read-only tools are exposed.

package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"mi-tech/internal/mcp"
	mcpRepoPkg "mi-tech/internal/mcp/repository"
	"mi-tech/internal/server"
	"mi-tech/internal/shared/config"
	"mi-tech/internal/shared/database"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	keyRepo := mcpRepoPkg.NewMachineKeyRepository(db)
	keyService := mcp.NewMachineKeyService(keyRepo)

	keyToken := os.Getenv("MCP_API_KEY")
	if keyToken == "" {
		slog.Error("MCP_API_KEY environment variable is required for stdio mode")
		os.Exit(1)
	}

	key, err := keyService.Authenticate(keyToken)
	if err != nil {
		slog.Error("Invalid MCP_API_KEY", "error", err)
		os.Exit(1)
	}
	if len(key.Scopes) == 0 {
		slog.Error("MCP_API_KEY has no scopes; nothing to expose")
		os.Exit(1)
	}

	// Build the full server graph (repositories, services, handlers) so the
	// read-only MCP mux has live backend dispatch, then serve it over stdio.
	srv := server.NewServer(cfg, db)
	sdkServer := srv.MCPServer(key.ID, key.Name, key.Scopes)
	if sdkServer == nil {
		slog.Error("MCP server not initialized")
		os.Exit(1)
	}

	visibleTools := 0
	for _, spec := range mcp.DefaultCatalog {
		for _, scope := range key.Scopes {
			if spec.Scope == scope {
				visibleTools++
				break
			}
		}
	}
	slog.Info("Starting MI Tech MCP server (stdio)", "scopes", key.Scopes, "tools", visibleTools)
	if err := mcp.ServeStdio(context.Background(), sdkServer); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		slog.Error("MCP stdio server error", "error", err)
		os.Exit(1)
	}
}
