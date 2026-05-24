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
	tableAllowlist  map[string]bool
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	// We use \b for word boundaries to avoid blocking things like "orders" (contains "order").
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	allowlist := map[string]bool{
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

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		tableAllowlist:  allowlist,
	}
}

// IsSafe checks if the SQL string contains any blocked mutation keywords or unauthorized table access.
func (g *QueryGuard) IsSafe(sql string) error {
	// Normalize whitespace
	normalized := strings.TrimSpace(strings.Join(strings.Fields(sql), " "))
	
	// Must start with SELECT
	if !strings.HasPrefix(strings.ToUpper(normalized), "SELECT") {
		return fmt.Errorf("SECURITY ALERT: only SELECT queries are allowed")
	}

	if g.blockedPatterns.MatchString(normalized) {
		// Log the first 100 characters of the blocked query for debugging
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// Table allowlist check
	tables := g.extractTables(normalized)
	for _, table := range tables {
		normalizedTable := g.normalizeTableName(table)
		if normalizedTable == "" {
			continue
		}
		if !g.tableAllowlist[normalizedTable] {
			return fmt.Errorf("SECURITY ALERT: unauthorized table access detected: %s", table)
		}
	}
	
	return nil
}

// extractTables identifies table names in FROM and JOIN clauses.
func (g *QueryGuard) extractTables(sql string) []string {
	var tables []string

	// Normalize spaces
	normalized := strings.Join(strings.Fields(sql), " ")
	words := strings.Fields(normalized)

	for i := 0; i < len(words)-1; i++ {
		word := strings.ToUpper(words[i])
		if word == "FROM" || word == "JOIN" {
			for j := i + 1; j < len(words); j++ {
				candidate := words[j]
				if strings.HasPrefix(candidate, "(") {
					// Subquery starts, skip '('
					continue
				}
				cleanName := strings.Trim(candidate, ",;()")
				if g.isKeyword(cleanName) {
					break
				}
				if cleanName != "" {
					tables = append(tables, cleanName)
				}
				// If it doesn't end with a comma, it's not a list of tables
				if !strings.HasSuffix(candidate, ",") {
					break
				}
			}
		}
	}

	return tables
}

func (g *QueryGuard) isKeyword(w string) bool {
	keywords := map[string]bool{
		"WHERE": true, "GROUP": true, "ORDER": true, "LIMIT": true, "HAVING": true,
		"UNION": true, "SELECT": true, "JOIN": true, "FROM": true, "ON": true, "USING": true,
		"INNER": true, "LEFT": true, "RIGHT": true, "FULL": true, "CROSS": true, "OUTER": true,
		"AS": true,
	}
	return keywords[strings.ToUpper(w)]
}

// normalizeTableName strips quotes and schema prefixes for allowlist matching.
func (g *QueryGuard) normalizeTableName(name string) string {
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "`", "")
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")

	// Strip schema prefix if present (e.g. public.orders -> orders)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}

	return strings.ToLower(strings.TrimSpace(name))
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
