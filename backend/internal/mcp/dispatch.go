package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
)

// MuxExecutor dispatches read-only tool calls to an internal mux that serves
// only GET routes from the read-only catalog. Requests execute in-process, so
// no network hops or additional authentication are required beyond the machine
// key already validated at the transport layer.
type MuxExecutor struct {
	handler      http.Handler
	writeHandler http.Handler
}

// identityExecutor binds a machine-key identity to an executor. This is used
// by the stdio transport, where there is no HTTP middleware to inject the
// authenticated key into the MCP request context.
type identityExecutor struct {
	inner  Executor
	id     int64
	name   string
	scopes []string
}

// WithMachineIdentity returns an executor that propagates machine-key
// identity into every dispatched request and audit record.
func WithMachineIdentity(inner Executor, id int64, name string, scopes []string) Executor {
	return &identityExecutor{inner: inner, id: id, name: name, scopes: scopes}
}

func (e *identityExecutor) Dispatch(ctx context.Context, tool ToolSpec, args map[string]any) (json.RawMessage, error) {
	ctx = context.WithValue(ctx, "machineKeyID", e.id)
	ctx = context.WithValue(ctx, "machineKeyName", e.name)
	ctx = context.WithValue(ctx, "machineScopes", e.scopes)
	return e.inner.Dispatch(ctx, tool, args)
}

// NewMuxExecutor creates an executor that routes tool calls to the given
// internal read-only handler (an *http.ServeMux).
func NewMuxExecutor(handler http.Handler) *MuxExecutor {
	return &MuxExecutor{handler: handler}
}

// NewMuxExecutorWithWriteHandler adds the explicitly scoped MCP write surface.
// The regular handler remains GET-only; only catalog tools marked Write can
// reach writeHandler.
func NewMuxExecutorWithWriteHandler(handler http.Handler, writeHandler http.Handler) *MuxExecutor {
	return &MuxExecutor{handler: handler, writeHandler: writeHandler}
}

// Dispatch builds a GET request for the tool and executes it against the
// internal mux. Tool arguments become query parameters (or path segments for
// tools declaring PathArgs). The machine-key identity is propagated into the
// request context so read-only handlers can resolve role/user context.
func (e *MuxExecutor) Dispatch(ctx context.Context, tool ToolSpec, args map[string]any) (json.RawMessage, error) {
	if err := normalizeArgs(tool, args); err != nil {
		return nil, err
	}

	path, err := buildToolPath(tool, args)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if !tool.Write {
		for _, a := range tool.Args {
			if inPath(tool.PathArgs, a.Name) {
				continue
			}
			v, ok := args[a.Name]
			if !ok || v == nil {
				continue
			}
			query.Set(a.Name, stringify(v))
		}
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	method := http.MethodGet
	requestHandler := e.handler
	var requestBody []byte
	if tool.Write {
		if e.writeHandler == nil {
			return nil, fmt.Errorf("tool %s has no write handler", tool.Name)
		}
		method = http.MethodPost
		requestHandler = e.writeHandler
		var err error
		writeArgs := make(map[string]any, len(args))
		for key, value := range args {
			if (key == "target_platforms" || key == "media_urls") && value != nil {
				if text, ok := value.(string); ok {
					items := make([]string, 0)
					for _, item := range strings.Split(text, ",") {
						if trimmed := strings.TrimSpace(item); trimmed != "" {
							items = append(items, trimmed)
						}
					}
					writeArgs[key] = items
					continue
				}
			}
			writeArgs[key] = value
		}
		requestBody, err = json.Marshal(writeArgs)
		if err != nil {
			return nil, ToolError{Status: http.StatusBadRequest, Err: fmt.Errorf("encode tool arguments: %w", err)}
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if tool.Write {
		req.Header.Set("Content-Type", "application/json")
	}
	req = req.WithContext(withIdentity(ctx, req.Context()))

	rec := httptest.NewRecorder()
	requestHandler.ServeHTTP(rec, req)

	if rec.Code >= 400 {
		return nil, ToolError{Status: rec.Code, Err: fmt.Errorf("tool %s failed: HTTP %d: %s", tool.Name, rec.Code, rec.Body.String())}
	}
	body := rec.Body.Bytes()
	if len(body) == 0 {
		body = []byte("{}")
	}
	return sanitizeResponse(body)
}

// buildToolPath assembles the request path, injecting path args (in declared
// order) after the route prefix.
func buildToolPath(tool ToolSpec, args map[string]any) (string, error) {
	path := strings.TrimRight(tool.Route, "/")
	for _, name := range tool.PathArgs {
		v, ok := args[name]
		if !ok || v == nil {
			return "", fmt.Errorf("missing required path argument %q for tool %s", name, tool.Name)
		}
		path += "/" + url.PathEscape(stringify(v))
	}
	return path, nil
}

func inPath(pathArgs []string, name string) bool {
	for _, p := range pathArgs {
		if p == name {
			return true
		}
	}
	return false
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

// withIdentity carries the machine-key identity into the request context so
// handlers that read user context (role, username, userID) behave correctly.
func withIdentity(mcpCtx, reqCtx context.Context) context.Context {
	ctx := reqCtx

	if id, ok := mcpCtx.Value("machineKeyID").(int64); ok {
		ctx = context.WithValue(ctx, "userID", id)
	}
	if name, ok := mcpCtx.Value("machineKeyName").(string); ok {
		ctx = context.WithValue(ctx, "username", name)
	}
	ctx = context.WithValue(ctx, "userRole", "read")

	// Preserve machine key scopes for downstream scope checks.
	if scopes, ok := mcpCtx.Value("machineScopes").([]string); ok {
		ctx = context.WithValue(ctx, "machineScopes", scopes)
	}
	return ctx
}
