package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Executor executes a single allowlisted tool call against the backend.
// It is the seam between the MCP protocol layer and the internal dispatch mux.
type Executor interface {
	// Dispatch runs the tool and returns its raw JSON result.
	Dispatch(ctx context.Context, tool ToolSpec, args map[string]any) (json.RawMessage, error)
}

// BuildServer constructs an MCP server from the catalog, registering every tool
// whose scope is present in allowedScopes. A nil allowedScopes enables all
// catalog tools (used only by the unrestricted local stdio entrypoint); a
// non-nil empty slice exposes no tools.
func BuildServer(catalog Catalog, executor Executor, audit *AuditService, allowedScopes []string) *sdk.Server {
	impl := &sdk.Implementation{Name: "mi-tech-mcp", Version: "0.1.0"}
	server := sdk.NewServer(impl, &sdk.ServerOptions{
		Capabilities: &sdk.ServerCapabilities{
			Tools: &sdk.ToolCapabilities{ListChanged: false},
		},
	})

	allowSet := map[string]struct{}{}
	if allowedScopes != nil {
		for _, s := range allowedScopes {
			allowSet[s] = struct{}{}
		}
	}

	for _, spec := range catalog {
		if allowedScopes != nil {
			if _, ok := allowSet[spec.Scope]; !ok {
				continue
			}
		}
		server.AddTool(&sdk.Tool{
			Name:        spec.Name,
			Description: spec.Description,
			InputSchema: inputSchema(spec),
		}, toolHandler(spec, executor, audit))
	}
	return server
}

// inputSchema builds a JSON Schema (2020-12 draft, as a plain map) from the
// tool's declared argument specs.
func inputSchema(spec ToolSpec) map[string]any {
	props := make(map[string]any, len(spec.Args))
	var required []string
	for _, a := range spec.Args {
		prop := map[string]any{"type": a.Type}
		if a.Description != "" {
			prop["description"] = a.Description
		}
		if a.Default != nil {
			prop["default"] = a.Default
		}
		props[a.Name] = prop
		if a.Required {
			required = append(required, a.Name)
		}
	}
	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// toolHandler wires a catalog tool to the executor and audit service.
func toolHandler(spec ToolSpec, executor Executor, audit *AuditService) sdk.ToolHandler {
	return func(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		start := time.Now()
		args := map[string]any{}
		if len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return auditedToolError(audit, ctx, spec, start, NewToolError(400, fmt.Sprintf("malformed tool arguments: %v", err)))
			}
		}

		data, err := executor.Dispatch(ctx, spec, args)

		if err != nil {
			return auditedToolError(audit, ctx, spec, start, err)
		}
		result := &sdk.CallToolResult{
			Content: []sdk.Content{&sdk.TextContent{Text: string(data)}},
		}
		if audit != nil {
			keyID, keyName, scopes := identityFromContext(ctx)
			// The in-process executor collapses successful backend responses to
			// the MCP result body, so 200 is the protocol-level success status.
			audit.Log(AuditEntry{KeyID: keyID, KeyName: keyName, Scopes: scopes, Tool: spec.Name, Outcome: "success", Status: http.StatusOK, DurationMs: time.Since(start).Milliseconds()})
		}
		return result, nil
	}
}

// ToolError preserves an internal/backend status while still returning a
// protocol-level MCP tool error instead of an HTTP 500.
type ToolError struct {
	Status int
	Err    error
}

func NewToolError(status int, message string) error {
	return ToolError{Status: status, Err: errors.New(message)}
}

func (e ToolError) Error() string { return e.Err.Error() }
func (e ToolError) Unwrap() error { return e.Err }

func auditedToolError(audit *AuditService, ctx context.Context, spec ToolSpec, start time.Time, err error) (*sdk.CallToolResult, error) {
	keyID, keyName, scopes := identityFromContext(ctx)
	status := 400
	var toolErr ToolError
	if errors.As(err, &toolErr) {
		status = toolErr.Status
	}
	if audit != nil {
		audit.Log(AuditEntry{KeyID: keyID, KeyName: keyName, Scopes: scopes, Tool: spec.Name, Outcome: "error", Status: status, DurationMs: time.Since(start).Milliseconds()})
	}
	return &sdk.CallToolResult{IsError: true, Content: []sdk.Content{&sdk.TextContent{Text: err.Error()}}}, nil
}

// StubExecutor is retained for protocol-layer tests and local scaffolding.
type StubExecutor struct{}

// Dispatch returns a clear "not yet implemented" result for any tool.
func (StubExecutor) Dispatch(ctx context.Context, tool ToolSpec, args map[string]any) (json.RawMessage, error) {
	_ = ctx
	_ = args
	return json.RawMessage(fmt.Sprintf(`{"error":"tool %q dispatch not yet implemented"}`, tool.Name)), fmt.Errorf("tool %q dispatch not yet implemented", tool.Name)
}

// identityFromContext extracts the machine-key identity (if present) from the
// request context. These keys match the values set by
// shared/middleware.MachineKeyMiddleware.
func identityFromContext(ctx context.Context) (id int64, name string, scopes []string) {
	if v, ok := ctx.Value("machineKeyID").(int64); ok {
		id = v
	}
	if v, ok := ctx.Value("machineKeyName").(string); ok {
		name = v
	}
	if v, ok := ctx.Value("machineScopes").([]string); ok {
		scopes = v
	}
	return id, name, scopes
}
