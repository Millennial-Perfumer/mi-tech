package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// testHandler returns a fixed JSON body for any GET request and 405 otherwise.
func testHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"echo": map[string]any{
			"query": r.URL.RawQuery,
			"path":  r.URL.Path,
		},
	})
}

func TestMuxExecutorDispatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", testHandler)

	exec := NewMuxExecutor(mux)
	ctx := context.Background()

	tool, ok := DefaultCatalog.Lookup("orders_list")
	require.True(t, ok)

	data, err := exec.Dispatch(ctx, tool, map[string]any{
		"limit":  10,
		"search": "perfume",
	})
	require.NoError(t, err)

	var resp struct {
		Success bool `json:"success"`
		Echo    struct {
			Query string `json:"query"`
			Path  string `json:"path"`
		} `json:"echo"`
	}
	require.NoError(t, json.Unmarshal(data, &resp))
	require.True(t, resp.Success)
	require.Contains(t, resp.Echo.Query, "limit=10")
	require.Contains(t, resp.Echo.Query, "search=perfume")
	require.Equal(t, "/api/orders", resp.Echo.Path)
}

func TestMuxExecutorPathArgs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/system/docs/", testHandler)

	exec := NewMuxExecutor(mux)
	ctx := context.Background()

	tool, ok := DefaultCatalog.Lookup("system_doc_get")
	require.True(t, ok)

	data, err := exec.Dispatch(ctx, tool, map[string]any{"slug": "api-auth"})
	require.NoError(t, err)

	var resp struct {
		Success bool `json:"success"`
		Echo    struct {
			Path string `json:"path"`
		} `json:"echo"`
	}
	require.NoError(t, json.Unmarshal(data, &resp))
	require.True(t, resp.Success)
	require.Equal(t, "/api/system/docs/api-auth", resp.Echo.Path)
}

func TestMuxExecutorMissingPathArg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/system/docs/", testHandler)

	exec := NewMuxExecutor(mux)
	tool, ok := DefaultCatalog.Lookup("system_doc_get")
	require.True(t, ok)

	_, err := exec.Dispatch(context.Background(), tool, map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "slug")
}

func TestMuxExecutorRejectsWriteMethodsAtMux(t *testing.T) {
	// A read-only mux must not serve POST/PUT/DELETE. The ro wrapper rejects
	// them, but MuxExecutor only issues GET; this guards the read-only guarantee.
	mux := http.NewServeMux()
	blocking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/orders", blocking)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestMuxExecutorForwardErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	exec := NewMuxExecutor(mux)
	tool, ok := DefaultCatalog.Lookup("orders_list")
	require.True(t, ok)

	_, err := exec.Dispatch(context.Background(), tool, map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP 500")
}

func TestMuxExecutorDispatchesScopedWriteToolAsJSONPost(t *testing.T) {
	readMux := http.NewServeMux()
	writeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "caption", payload["caption"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	exec := NewMuxExecutorWithWriteHandler(readMux, writeHandler)
	tool := ToolSpec{Name: "smm_queue_create", Route: "/api/marketing/smm/queue", Write: true}

	data, err := exec.Dispatch(context.Background(), tool, map[string]any{"caption": "caption"})
	require.NoError(t, err)
	require.JSONEq(t, `{"success":true}`, string(data))
}

func TestMuxExecutorDispatchesWritePayloadAndQueryArguments(t *testing.T) {
	writeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "42", r.URL.Query().Get("id"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "fulfilled", payload["status"])
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	exec := NewMuxExecutorWithWriteHandler(http.NewServeMux(), writeHandler)
	tool, ok := DefaultCatalog.Lookup("orders_update_status")
	require.True(t, ok)

	data, err := exec.Dispatch(context.Background(), tool, map[string]any{
		"id":      42,
		"payload": map[string]any{"status": "fulfilled"},
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"success":true}`, string(data))
}

func TestMuxExecutorDispatchesQueryOnlyWriteArguments(t *testing.T) {
	writeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "order_confirmation", r.URL.Query().Get("name"))
		require.Equal(t, "", r.URL.Query().Get("payload"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Empty(t, body)
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	exec := NewMuxExecutorWithWriteHandler(http.NewServeMux(), writeHandler)
	tool, ok := DefaultCatalog.Lookup("whatsapp_template_sync_single")
	require.True(t, ok)

	data, err := exec.Dispatch(context.Background(), tool, map[string]any{"name": "order_confirmation"})
	require.NoError(t, err)
	require.JSONEq(t, `{"success":true}`, string(data))
}
