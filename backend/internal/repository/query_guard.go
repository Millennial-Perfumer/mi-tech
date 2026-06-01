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
	fromClauseRegex *regexp.Regexp
	joinClauseRegex *regexp.Regexp
	allowedTables   map[string]bool
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	// Extracts the content after FROM until the next major keyword (WHERE, GROUP, ORDER, etc.)
	// Handles comma-separated tables.
	fromPattern := `(?i)\bFROM\s+([^;]+?)(?:\s+(?:WHERE|GROUP|ORDER|LIMIT|JOIN|HAVING|UNION|INTERSECT|EXCEPT)|$)`

	// Extracts table names after JOIN.
	joinPattern := `(?i)\bJOIN\s+([a-zA-Z0-9_.]+)`

	// Strict allowlist of tables the AI is permitted to query.
	allowed := map[string]bool{
		"orders":                 true,
		"order_line_items":       true,
		"customers":              true,
		"inventory_items":        true,
		"sources":                true,
		"webhook_events":         true,
		"oil_inventory":          true,
		"manufacturing_records":  true,
		"manufacturing_products": true,
		"purchase_orders":        true,
		"feedback":               true,
		"planner_tasks":          true,
	}

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		fromClauseRegex: regexp.MustCompile(fromPattern),
		joinClauseRegex: regexp.MustCompile(joinPattern),
		allowedTables:   allowed,
	}
}

// IsSafe checks if the SQL string contains any blocked mutation keywords and only accesses allowed tables.
func (g *QueryGuard) IsSafe(sql string) error {
	// Normalize whitespace
	normalized := strings.TrimSpace(strings.Join(strings.Fields(sql), " "))
	
	// Must start with SELECT
	if !strings.HasPrefix(strings.ToUpper(normalized), "SELECT") {
		return fmt.Errorf("SECURITY ALERT: only SELECT queries are allowed")
	}

	// 1. Check for mutation keywords
	if g.blockedPatterns.MatchString(normalized) {
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// 2. Check for unauthorized table access in FROM clauses
	fromMatches := g.fromClauseRegex.FindAllStringSubmatch(normalized, -1)
	for _, match := range fromMatches {
		if len(match) > 1 {
			// Content of the FROM clause, e.g. "orders, users" or "orders o"
			content := match[1]
			// Split by comma for multiple tables
			tables := strings.Split(content, ",")
			for _, t := range tables {
				if err := g.validateTableString(t); err != nil {
					return err
				}
			}
		}
	}

	// 3. Check for unauthorized table access in JOIN clauses
	joinMatches := g.joinClauseRegex.FindAllStringSubmatch(normalized, -1)
	for _, match := range joinMatches {
		if len(match) > 1 {
			if err := g.validateTableString(match[1]); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateTableString cleans up a table string (removes aliases and whitespace) and checks the allowlist.
func (g *QueryGuard) validateTableString(t string) error {
	t = strings.TrimSpace(t)
	if t == "" {
		return nil
	}

	// Detect subqueries (contains parentheses)
	if strings.Contains(t, "(") {
		return fmt.Errorf("SECURITY ALERT: subqueries in FROM/JOIN are restricted for AI analysis")
	}

	// Remove aliases (take the first word)
	parts := strings.Fields(t)
	if len(parts) == 0 {
		return nil
	}
	tableName := strings.ToLower(parts[0])

	// Handle schema prefix if present (e.g., public.orders)
	if parts := strings.Split(tableName, "."); len(parts) > 1 {
		tableName = parts[len(parts)-1]
	}

	if !g.allowedTables[tableName] {
		return fmt.Errorf("SECURITY ALERT: access to table '%s' is restricted", tableName)
	}
	return nil
}

// IsTableAllowed checks if a specific table name is in the allowlist.
func (g *QueryGuard) IsTableAllowed(tableName string) bool {
	tableName = strings.ToLower(tableName)
	if parts := strings.Split(tableName, "."); len(parts) > 1 {
		tableName = parts[len(parts)-1]
	}
	return g.allowedTables[tableName]
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
