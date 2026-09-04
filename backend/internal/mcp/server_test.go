package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"mi-tech/internal/mcp/entity"
)

type successfulTestExecutor struct{}

func (successfulTestExecutor) Dispatch(context.Context, ToolSpec, map[string]any) (json.RawMessage, error) {
	return json.RawMessage(`{"success":true}`), nil
}

type captureAuditRepository struct {
	entries chan *entity.MCPAuditLog
}

func (r *captureAuditRepository) Create(entry *entity.MCPAuditLog) error {
	r.entries <- entry
	return nil
}

// connectClient connects an SDK client to a built server over in-memory
// transports and returns the client session.
func connectClient(t *testing.T, server *sdk.Server) *sdk.ClientSession {
	t.Helper()
	ctx := context.Background()

	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		session.Close()
		ss.Close()
	})
	return session
}

// TestBuildServerRegistersAllTools verifies the full catalog is discoverable.
func TestBuildServerRegistersAllTools(t *testing.T) {
	server := BuildServer(DefaultCatalog, StubExecutor{}, nil, nil)
	session := connectClient(t, server)

	var names []string
	for tool, err := range session.Tools(context.Background(), nil) {
		require.NoError(t, err)
		names = append(names, tool.Name)
	}
	require.Len(t, names, len(DefaultCatalog))
}

// TestBuildServerScopesTools verifies per-scope tool registration filters tools.
func TestBuildServerScopesTools(t *testing.T) {
	server := BuildServer(DefaultCatalog, StubExecutor{}, nil, []string{ScopeOrders, ScopeMetrics})
	session := connectClient(t, server)

	var names []string
	for tool, err := range session.Tools(context.Background(), nil) {
		require.NoError(t, err)
		names = append(names, tool.Name)
	}
	require.Contains(t, names, "orders_list")
	require.Contains(t, names, "dashboard_metrics")
	require.NotContains(t, names, "gst_summary")
	require.NotContains(t, names, "support_tickets")
}

func TestBuildServerEmptyScopesExposeNoTools(t *testing.T) {
	server := BuildServer(DefaultCatalog, StubExecutor{}, nil, []string{})
	session := connectClient(t, server)

	count := 0
	for tool, err := range session.Tools(context.Background(), nil) {
		require.NoError(t, err)
		require.NotEmpty(t, tool.Name)
		count++
	}
	require.Zero(t, count)
}

func TestToolHandlerRejectsMalformedArguments(t *testing.T) {
	spec, ok := DefaultCatalog.Lookup("orders_list")
	require.True(t, ok)
	handler := toolHandler(spec, StubExecutor{}, nil)

	result, err := handler(context.Background(), &sdk.CallToolRequest{
		Params: &sdk.CallToolParamsRaw{Arguments: json.RawMessage(`{"limit":`)},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Contains(t, result.Content[0].(*sdk.TextContent).Text, "malformed tool arguments")
}

// TestToolCallReturnsStubError verifies a tool call is routed to the executor.
func TestToolCallReturnsStubError(t *testing.T) {
	server := BuildServer(DefaultCatalog, StubExecutor{}, nil, nil)
	session := connectClient(t, server)

	res, err := session.CallTool(context.Background(), &sdk.CallToolParams{
		Name:      "orders_list",
		Arguments: map[string]any{"limit": 10},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "stub executor should report an error")
}

func TestToolHandlerAuditsSuccessfulCalls(t *testing.T) {
	repo := &captureAuditRepository{entries: make(chan *entity.MCPAuditLog, 1)}
	audit := NewAuditService(repo, 1)
	t.Cleanup(audit.Close)

	spec, ok := DefaultCatalog.Lookup("orders_list")
	require.True(t, ok)
	ctx := context.WithValue(context.Background(), "machineKeyID", int64(7))
	ctx = context.WithValue(ctx, "machineKeyName", "test-key")
	ctx = context.WithValue(ctx, "machineScopes", []string{ScopeOrders})

	result, err := toolHandler(spec, successfulTestExecutor{}, audit)(ctx, &sdk.CallToolRequest{
		Params: &sdk.CallToolParamsRaw{Arguments: json.RawMessage(`{}`)},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	select {
	case entry := <-repo.entries:
		require.Equal(t, "orders_list", entry.Tool)
		require.Equal(t, "success", entry.Outcome)
		require.Equal(t, 200, entry.Status)
		require.Equal(t, int64(7), entry.KeyID)
	case <-time.After(time.Second):
		t.Fatal("successful tool call was not audited")
	}
}

// TestToolInputSchema verifies generated input schemas carry args.
func TestToolInputSchema(t *testing.T) {
	spec, ok := DefaultCatalog.Lookup("orders_list")
	require.True(t, ok)

	schema := inputSchema(spec)
	require.Equal(t, "object", schema["type"])
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, props, "start_date")
	require.Contains(t, props, "limit")

	queueSpec, ok := DefaultCatalog.Lookup("smm_queue_create")
	require.True(t, ok)
	queueSchema := inputSchema(queueSpec)
	queueProps, ok := queueSchema["properties"].(map[string]any)
	require.True(t, ok)
	mediaFiles, ok := queueProps["media_files"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "array", mediaFiles["type"])
	items, ok := mediaFiles["items"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "string", items["type"])
}

// TestHTTPHandlerStateless verifies the Streamable HTTP handler mounts.
func TestHTTPHandlerStateless(t *testing.T) {
	handler := HTTPHandler(func(scopes []string) *sdk.Server {
		return BuildServer(DefaultCatalog, StubExecutor{}, nil, scopes)
	}, HTTPHandlerOptions{Stateless: true})
	require.NotNil(t, handler)
}
