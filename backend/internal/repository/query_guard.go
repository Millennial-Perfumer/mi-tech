package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryGuard provides runtime safety by blocking SQL mutation keywords
// and enforcing a table allowlist.
type QueryGuard struct {
	blockedPatterns *regexp.Regexp
	allowedTables   map[string]bool
	clauseRegex     *regexp.Regexp
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	// Allowed tables for AI analysis.
	allowed := map[string]bool{
		"orders":                 true,
		"order_line_items":       true,
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
	}

	// Clause regex to find potential table names after FROM or JOIN
	// We use a pattern that captures the content after FROM or JOIN
	clauseRegex := regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+([a-z0-9_".\s,]+)`)

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		allowedTables:   allowed,
		clauseRegex:     clauseRegex,
	}
}

// IsSafe checks if the SQL string is a safe SELECT query.
func (g *QueryGuard) IsSafe(sql string) error {
	// Normalize whitespace
	normalized := strings.TrimSpace(strings.Join(strings.Fields(sql), " "))
	
	// Must start with SELECT
	if !strings.HasPrefix(strings.ToUpper(normalized), "SELECT") {
		return fmt.Errorf("SECURITY ALERT: only SELECT queries are allowed")
	}

	// Block mutation keywords
	if g.blockedPatterns.MatchString(normalized) {
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// Recursive check for subqueries in parentheses
	subQueryRegex := regexp.MustCompile(`(?i)\((SELECT[^)]+)\)`)
	subMatches := subQueryRegex.FindAllStringSubmatch(normalized, -1)
	for _, subMatch := range subMatches {
		if err := g.IsSafe(subMatch[1]); err != nil {
			return err
		}
	}

	// Enforce table allowlist
	tables := g.extractTables(normalized)
	for _, table := range tables {
		cleanTable := g.normalizeTableName(table)
		if cleanTable == "" {
			continue
		}
		if !g.allowedTables[cleanTable] {
			return fmt.Errorf("SECURITY ALERT: access to table '%s' is restricted", cleanTable)
		}
	}
	
	return nil
}

// extractTables finds potential table names in the query.
func (g *QueryGuard) extractTables(sql string) []string {
	var tables []string

	// Temporarily remove subqueries for simpler processing
	depth := 0
	var sb strings.Builder
	for _, char := range sql {
		if char == '(' {
			depth++
		}
		if depth == 0 {
			sb.WriteRune(char)
		}
		if char == ')' {
			depth--
		}
	}
	cleanSQL := sb.String()

	// Tokenize the SQL by spaces to iterate through words
	words := strings.Fields(cleanSQL)
	for i := 0; i < len(words)-1; i++ {
		word := strings.ToUpper(words[i])
		if word == "FROM" || word == "JOIN" {
			// The next part contains table names, potentially comma-separated
			// We look at all words until we hit another keyword
			for j := i + 1; j < len(words); j++ {
				content := words[j]
				if isSQLKeyword(content) {
					break
				}

				// Handle comma-separated tables within this word or spanning multiple words
				commaParts := strings.Split(content, ",")
				for _, cp := range commaParts {
					if cp == "" {
						continue
					}
					// Only the first part of a word is the table, the rest might be alias
					// But if it was split by comma, each part might be a table
					// Actually, if we have "orders,customers", both are tables.
					// If we have "orders o", only "orders" is the table.
					tables = append(tables, cp)
				}

				// If this word didn't end in a comma, and the next word isn't a keyword,
				// then this word might be a table and the next an alias, OR the next is another table via comma.
				// This is getting complex. Let's simplify.

				// If word doesn't end in comma and we have more words
				if !strings.HasSuffix(content, ",") {
					break // Assume next word is alias or next clause
				}
			}
		}
	}

	return tables
}

// isSQLKeyword checks if a word is a SQL keyword that might follow a table name.
func isSQLKeyword(s string) bool {
	k := strings.ToUpper(s)
	switch k {
	case "ON", "WHERE", "GROUP", "ORDER", "LIMIT", "LEFT", "RIGHT", "INNER", "OUTER", "UNION", "HAVING", "OFFSET", "SELECT":
		return true
	}
	return false
}

// normalizeTableName removes quotes and schema prefixes from a table name.
func (g *QueryGuard) normalizeTableName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "\"", "")
	if parts := strings.Split(name, "."); len(parts) > 1 {
		name = parts[len(parts)-1]
	}
	return strings.TrimSpace(name)
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
