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
	fromJoinRegex   *regexp.Regexp
	allowedTables   map[string]bool
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	mutationPattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	// Regex to find the start of FROM or JOIN clauses.
	fromJoinPattern := `(?i)\b(FROM|JOIN)\s+`

	allowed := map[string]bool{
		"orders":                true,
		"customers":             true,
		"inventory_items":       true,
		"order_line_items":      true,
		"manufacturing_records": true,
		"oil_inventory":         true,
	}

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(mutationPattern),
		fromJoinRegex:   regexp.MustCompile(fromJoinPattern),
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

	// Check for mutation keywords
	if g.blockedPatterns.MatchString(normalized) {
		snippet := normalized
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		return fmt.Errorf("SECURITY ALERT: mutation keyword detected in AI query: %s", snippet)
	}

	// Check table allowlist
	// Find all occurrences of FROM/JOIN
	indices := g.fromJoinRegex.FindAllStringIndex(normalized, -1)
	for _, idx := range indices {
		// content starts after FROM/JOIN
		content := normalized[idx[1]:]

		// Stop at the next keyword
		stopKeywords := []string{" WHERE ", " GROUP ", " ORDER ", " LIMIT ", " OFFSET ", " ON ", " USING ", " HAVING ", " WINDOW ", " FOR ", " LOCK ", " JOIN ", " FROM "}
		upperContent := " " + strings.ToUpper(content) + " "
		earliestStop := len(content)

		for _, sk := range stopKeywords {
			if stopIdx := strings.Index(upperContent, sk); stopIdx != -1 {
				if stopIdx < earliestStop {
					earliestStop = stopIdx
				}
			}
		}

		tablePart := content[:earliestStop]

		// Split by comma for multiple tables in FROM
		tables := strings.Split(tablePart, ",")
		for _, t := range tables {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}

			// Table name is the first word (handles aliases)
			tableNameWithSchema := strings.Fields(t)[0]

			// Handle schema qualification and quotes
			parts := strings.Split(tableNameWithSchema, ".")
			tableName := parts[len(parts)-1]
			tableName = strings.Trim(tableName, `"'`)
			tableName = strings.ToLower(tableName)

			if !g.allowedTables[tableName] {
				return fmt.Errorf("SECURITY ALERT: access to table '%s' is not allowed", tableName)
			}
		}
	}

	// Also check for subqueries by recursively calling IsSafe on content inside parentheses
	content := normalized
	for {
		start := strings.Index(strings.ToUpper(content), "(SELECT ")
		if start == -1 {
			break
		}

		// Find matching closing parenthesis
		bracketCount := 0
		end := -1
		for i := start; i < len(content); i++ {
			if content[i] == '(' {
				bracketCount++
			} else if content[i] == ')' {
				bracketCount--
				if bracketCount == 0 {
					end = i
					break
				}
			}
		}

		if end != -1 {
			subquery := content[start+1 : end]
			if err := g.IsSafe(subquery); err != nil {
				return err
			}
			// Continue after this subquery
			content = content[end+1:]
		} else {
			break
		}
	}
	
	return nil
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
