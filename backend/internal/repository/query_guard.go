package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryGuard provides runtime safety by blocking SQL mutation keywords and enforcing a table allowlist.
// This is a defense-in-depth layer for AI-generated or AI-triggered queries.
type QueryGuard struct {
	blockedPatterns *regexp.Regexp
	allowedTables   map[string]bool
	tableRegex      *regexp.Regexp
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	allowed := map[string]bool{
		"orders":                 true,
		"customers":              true,
		"inventory_items":        true,
		"inventory_mappings":     true,
		"inventory_logs":         true,
		"suppliers":              true,
		"oil_inventory":          true,
		"purchase_orders":        true,
		"manufacturing_records":  true,
		"manufacturing_oils":     true,
		"manufacturing_products": true,
		"planner_boards":         true,
		"planner_columns":        true,
		"planner_sprints":        true,
		"planner_tasks":          true,
		"planner_task_logs":      true,
		"sources":                true,
		"feedback_statuses":      true,
		"customer_feedback":      true,
		"order_line_items":       true,
	}

	// Regex to find table names after FROM or JOIN.
	// It handles: FROM table, JOIN table, FROM "table", FROM public.table
	tableRegex := regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+([a-z0-9_".]+)`)

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		allowedTables:   allowed,
		tableRegex:      tableRegex,
	}
}

// IsSafe checks if the SQL string is a SELECT query, contains no mutation keywords,
// and only references allowed tables.
func (g *QueryGuard) IsSafe(sql string) error {
	// Normalize whitespace
	normalized := strings.TrimSpace(strings.Join(strings.Fields(sql), " "))

	// Must start with SELECT
	if !strings.HasPrefix(strings.ToUpper(normalized), "SELECT") {
		return fmt.Errorf("SECURITY ALERT: only SELECT queries are allowed")
	}

	// Check for blocked mutation keywords
	if g.blockedPatterns.MatchString(normalized) {
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// Extract and validate table names
	matches := g.tableRegex.FindAllStringSubmatch(normalized, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		tableName := g.normalizeTableName(match[1])
		if !g.allowedTables[tableName] {
			return fmt.Errorf("SECURITY ALERT: unauthorized table access detected: %s", tableName)
		}
	}

	return nil
}

// normalizeTableName removes quotes and schema prefixes (e.g. "public.orders" -> "orders")
func (g *QueryGuard) normalizeTableName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "\"", "")

	// If it has a schema prefix, take the last part
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
