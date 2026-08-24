package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Validation limits applied to tool arguments (Phase 5 normalization).
const (
	DefaultPageSize = 50
	MaxPageSize     = 500
)

// normalizeArgs validates tool arguments before dispatch:
//   - date ranges must be valid ISO dates and start <= end
//   - pagination args are bounded to sane limits
//
// It mutates args in place (applying defaults for omitted pagination).
func normalizeArgs(tool ToolSpec, args map[string]any) error {
	if err := validateDateRange(tool, args); err != nil {
		return err
	}
	applyPaginationDefaults(tool, args)
	return nil
}

// validateDateRange rejects malformed dates and reversed ranges.
func validateDateRange(tool ToolSpec, args map[string]any) error {
	var start, end string
	for _, a := range tool.Args {
		switch a.Name {
		case "start_date", "startDate":
			start = stringValue(args[a.Name])
		case "end_date", "endDate":
			end = stringValue(args[a.Name])
		}
	}
	if start == "" && end == "" {
		return nil
	}

	var startT, endT time.Time
	var err error
	if start != "" {
		startT, err = parseISODate(start)
		if err != nil {
			return fmt.Errorf("invalid start_date %q for tool %s: must be YYYY-MM-DD", start, tool.Name)
		}
	}
	if end != "" {
		endT, err = parseISODate(end)
		if err != nil {
			return fmt.Errorf("invalid end_date %q for tool %s: must be YYYY-MM-DD", end, tool.Name)
		}
	}
	if start != "" && end != "" && startT.After(endT) {
		return fmt.Errorf("start_date %q is after end_date %q for tool %s", start, end, tool.Name)
	}
	return nil
}

// parseISODate parses a YYYY-MM-DD date.
func parseISODate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// applyPaginationDefaults bounds limit/page/offset for list tools.
func applyPaginationDefaults(tool ToolSpec, args map[string]any) {
	for _, a := range tool.Args {
		switch a.Name {
		case "limit":
			if v, ok := args["limit"]; ok {
				if n, ok := toInt(v); ok {
					if n <= 0 {
						args["limit"] = DefaultPageSize
					} else if n > MaxPageSize {
						args["limit"] = MaxPageSize
					}
				} else {
					args["limit"] = DefaultPageSize
				}
			} else {
				args["limit"] = DefaultPageSize
			}
		case "offset":
			if v, ok := args["offset"]; ok {
				if n, ok := toInt(v); ok && n > 0 {
					args["offset"] = n
				} else {
					args["offset"] = 0
				}
			}
		case "page":
			if v, ok := args["page"]; ok {
				if n, ok := toInt(v); ok && n > 0 {
					args["page"] = n
				} else {
					args["page"] = 1
				}
			}
		case "pageSize":
			if v, ok := args["pageSize"]; ok {
				if n, ok := toInt(v); ok && n > 0 {
					if n > MaxPageSize {
						args["pageSize"] = MaxPageSize
					} else {
						args["pageSize"] = n
					}
				} else {
					args["pageSize"] = DefaultPageSize
				}
			}
		}
	}
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if n, ok := v.(json.Number); ok {
		return n.String()
	}
	if f, ok := v.(float64); ok {
		return fmt.Sprintf("%.0f", f)
	}
	return ""
}

func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case json.Number:
		n, err := t.Int64()
		return int(n), err == nil
	case int:
		return t, true
	case string:
		var n int
		if _, err := fmt.Sscanf(t, "%d", &n); err == nil {
			return n, true
		}
	}
	return 0, false
}

// sensitiveField reports whether a JSON object field holds sensitive customer or
// configuration data that must be masked.
func sensitiveField(key string) bool {
	k := strings.ToLower(key)
	if strings.Contains(k, "phone") || strings.Contains(k, "email") {
		return true
	}
	if strings.Contains(k, "address") || strings.Contains(k, "pin") || strings.Contains(k, "zip") {
		return true
	}
	if strings.Contains(k, "secret") || strings.Contains(k, "token") || strings.Contains(k, "password") {
		return true
	}
	return false
}

// sanitizeResponse recursively walks a tool result and masks sensitive fields.
// Masking is lossy but preserves shape so clients can still reason about the data.
func sanitizeResponse(raw json.RawMessage) (json.RawMessage, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		// Not JSON; return as-is (e.g. plain-text docs).
		return raw, nil
	}
	sanitizeValue(&v)
	out, err := json.Marshal(v)
	if err != nil {
		return raw, nil
	}
	return out, nil
}

func sanitizeValue(v *any) {
	switch t := (*v).(type) {
	case map[string]any:
		for k, val := range t {
			if sensitiveField(k) {
				switch val.(type) {
				case string:
					if s := val.(string); s != "" {
						t[k] = maskValue(s)
					}
				}
				continue
			}
			sanitizeValue(&val)
			t[k] = val
		}
	case []any:
		for i := range t {
			sanitizeValue(&t[i])
		}
	}
}

// maskValue partially masks a sensitive string, keeping the first 2 and last 2
// characters visible. Short values are fully masked.
func maskValue(s string) string {
	r := []rune(s)
	if len(r) <= 4 {
		return "••••"
	}
	head := string(r[:2])
	tail := string(r[len(r)-2:])
	return head + "••••" + tail
}
