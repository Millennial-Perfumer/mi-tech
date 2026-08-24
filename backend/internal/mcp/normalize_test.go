package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeArgsDateRangeValid(t *testing.T) {
	tool, _ := DefaultCatalog.Lookup("dashboard_metrics")
	args := map[string]any{"start_date": "2026-08-01", "end_date": "2026-08-23"}
	require.NoError(t, normalizeArgs(tool, args))
}

func TestNormalizeArgsDateRangeReversed(t *testing.T) {
	tool, _ := DefaultCatalog.Lookup("dashboard_metrics")
	args := map[string]any{"start_date": "2026-08-23", "end_date": "2026-08-01"}
	err := normalizeArgs(tool, args)
	require.Error(t, err)
	require.Contains(t, err.Error(), "after")
}

func TestNormalizeArgsDateRangeMalformed(t *testing.T) {
	tool, _ := DefaultCatalog.Lookup("dashboard_metrics")
	args := map[string]any{"start_date": "23/08/2026"}
	err := normalizeArgs(tool, args)
	require.Error(t, err)
	require.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestNormalizeArgsPaginationDefaults(t *testing.T) {
	tool, _ := DefaultCatalog.Lookup("orders_list")
	args := map[string]any{}
	require.NoError(t, normalizeArgs(tool, args))
	require.Equal(t, DefaultPageSize, args["limit"])

	args = map[string]any{"limit": 100000}
	require.NoError(t, normalizeArgs(tool, args))
	require.Equal(t, MaxPageSize, args["limit"])

	args = map[string]any{"limit": -5}
	require.NoError(t, normalizeArgs(tool, args))
	require.Equal(t, DefaultPageSize, args["limit"])
}

func TestSanitizeResponseMasksSensitiveFields(t *testing.T) {
	raw := json.RawMessage(`{
		"success": true,
		"orders": [{
			"id": 1,
			"customer_name": "Jane Doe",
			"customer_phone": "919876543210",
			"customer_email": "jane@example.com",
			"customer_address1": "123 Main St",
			"customer_zip": "560001",
			"total_price": "100.00"
		}]
	}`)
	out, err := sanitizeResponse(raw)
	require.NoError(t, err)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(out, &resp))
	orders := resp["orders"].([]any)
	first := orders[0].(map[string]any)

	require.Equal(t, "Jane Doe", first["customer_name"]) // name not sensitive
	require.NotEqual(t, "919876543210", first["customer_phone"])
	require.NotEqual(t, "jane@example.com", first["customer_email"])
	require.NotEqual(t, "123 Main St", first["customer_address1"])
	require.NotEqual(t, "560001", first["customer_zip"])
	require.Equal(t, "100.00", first["total_price"])
}

func TestSanitizeResponseKeepsNonJSON(t *testing.T) {
	raw := json.RawMessage("# markdown doc\ncontent")
	out, err := sanitizeResponse(raw)
	require.NoError(t, err)
	require.Equal(t, string(raw), string(out))
}

func TestSanitizeResponsePreservesShape(t *testing.T) {
	raw := json.RawMessage(`{"success":true,"items":[{"sku":"A1","price":10}]}`)
	out, err := sanitizeResponse(raw)
	require.NoError(t, err)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(out, &resp))
	require.Equal(t, true, resp["success"])
	require.Len(t, resp["items"].([]any), 1)
}

func TestMaskValue(t *testing.T) {
	require.Equal(t, "ja••••om", maskValue("jane@example.com"))
	require.Equal(t, "••••", maskValue("abc"))
}
