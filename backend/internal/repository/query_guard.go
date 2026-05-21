package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryGuard provides runtime safety by blocking SQL mutation keywords.
// This is a defense-in-depth layer for AI-generated or AI-triggered queries.
type QueryGuard struct {
	blockedPatterns *regexp.Regexp
	allowedTables   map[string]bool
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	// We use \b for word boundaries to avoid blocking things like "orders" (contains "order").
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	// allowedTables defines a strict allowlist of tables that AI is permitted to query.
	// Sensitive tables like 'users', 'app_configs', 'webhooks' etc. are strictly excluded.
	allowedTables := map[string]bool{
		"orders":                 true,
		"order_line_items":       true,
		"customers":              true,
		"inventory_items":        true,
		"inventory_logs":         true,
		"inventory_mappings":     true,
		"suppliers":              true,
		"oil_inventory":          true,
		"manufacturing_records":  true,
		"manufacturing_oils":     true,
		"manufacturing_products": true,
		"automation_templates":   true,
		"social_metrics":         true,
		"customer_feedback":      true,
		"planner_boards":         true,
		"planner_columns":        true,
		"planner_tasks":          true,
	}

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		allowedTables:   allowedTables,
	}
}

// IsSafe checks if the SQL string contains any blocked mutation keywords
// and ensures only allowed tables are referenced.
func (g *QueryGuard) IsSafe(sql string) error {
	// Normalize whitespace
	normalized := strings.TrimSpace(strings.Join(strings.Fields(sql), " "))
	upperSQL := strings.ToUpper(normalized)

	// Must start with SELECT
	if !strings.HasPrefix(upperSQL, "SELECT") {
		return fmt.Errorf("SECURITY ALERT: only SELECT queries are allowed")
	}

	if g.blockedPatterns.MatchString(normalized) {
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// Enhanced table allowlist validation.
	// This covers 'FROM table', 'JOIN table', and comma-separated 'FROM table1, table2'.
	// It also handles quoted identifiers and schema prefixes like 'public.table'.

	tokens := strings.Fields(normalized)
	for i, token := range tokens {
		upperToken := strings.ToUpper(token)
		if upperToken == "FROM" || upperToken == "JOIN" {
			if i+1 < len(tokens) {
				for j := i + 1; j < len(tokens); j++ {
					subToken := tokens[j]
					upperSub := strings.ToUpper(subToken)
					if isKeyword(upperSub) {
						break
					}

					refs := strings.Split(subToken, ",")
					for _, ref := range refs {
						ref = strings.TrimSpace(ref)
						if ref == "" {
							continue
						}

						// The first part is the table name
						tableName := strings.Split(ref, " ")[0]
						// Strip ALL quotes first
						tableName = strings.ReplaceAll(tableName, `"`, "")
						tableName = strings.ReplaceAll(tableName, `'`, "")
						tableName = strings.ReplaceAll(tableName, "`", "")

						if dotIndex := strings.LastIndex(tableName, "."); dotIndex != -1 {
							tableName = tableName[dotIndex+1:]
						}
						tableName = strings.ToLower(tableName)

						if tableName != "" && !g.allowedTables[tableName] {
							return fmt.Errorf("SECURITY ALERT: access to table '%s' is not allowed", tableName)
						}
					}

					if !strings.Contains(subToken, ",") && j > i+1 {
						// Likely alias or ON clause
					}
				}
			}
		}
	}

	return nil
}

func isKeyword(s string) bool {
	keywords := map[string]bool{
		"WHERE": true, "GROUP": true, "ORDER": true, "LIMIT": true,
		"OFFSET": true, "FETCH": true, "UNION": true, "INTERSECT": true,
		"EXCEPT": true, "ON": true, "USING": true, "LEFT": true,
		"RIGHT": true, "INNER": true, "OUTER": true, "CROSS": true,
		"FULL": true, "JOIN": true,
	}
	return keywords[s]
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
