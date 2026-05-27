package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryGuard provides runtime safety by blocking SQL mutation keywords
// and enforcing a strict table allowlist for AI-generated or AI-triggered queries.
type QueryGuard struct {
	blockedPatterns *regexp.Regexp
	tableAllowlist  map[string]bool
	fromJoinRegex   *regexp.Regexp
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	// We use \b for word boundaries to avoid blocking things like "orders" (contains "order").
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME|INTO|COMMENT)\b`

	// Permitted tables for business data analysis.
	// Sensitive tables like 'users' and 'app_configs' are strictly excluded.
	allowlist := map[string]bool{
		"orders":               true,
		"order_line_items":     true,
		"customers":            true,
		"inventory_items":      true,
		"purchase_orders":      true,
		"purchase_order_items": true,
		"oil_inventory":       true,
		"suppliers":            true,
		"manufacturing_orders": true,
		"manufacturing_items":  true,
		"inventory_logs":       true,
		"automation_templates": true,
	}

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		tableAllowlist:  allowlist,
		fromJoinRegex:   regexp.MustCompile(`(?i)\b(FROM|JOIN)\s+([a-zA-Z0-9_]+)`),
	}
}

// IsSafe checks if the SQL string contains any blocked mutation keywords
// or references any tables not in the allowlist.
func (g *QueryGuard) IsSafe(sql string) error {
	// Normalize whitespace
	normalized := strings.TrimSpace(strings.Join(strings.Fields(sql), " "))
	
	// Must start with SELECT
	if !strings.HasPrefix(strings.ToUpper(normalized), "SELECT") {
		return fmt.Errorf("SECURITY ALERT: only SELECT queries are allowed")
	}

	// 1. Check for blocked mutation keywords
	if g.blockedPatterns.MatchString(normalized) {
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// 2. Check table allowlist
	matches := g.fromJoinRegex.FindAllStringSubmatch(normalized, -1)
	for _, match := range matches {
		if len(match) > 2 {
			tableName := strings.ToLower(match[2])
			if !g.IsTableAllowed(tableName) {
				return fmt.Errorf("SECURITY ALERT: unauthorized access to table '%s'", tableName)
			}
		}
	}
	
	return nil
}

// IsTableAllowed returns true if the table is in the permitted allowlist.
func (g *QueryGuard) IsTableAllowed(tableName string) bool {
	return g.tableAllowlist[strings.ToLower(tableName)]
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
