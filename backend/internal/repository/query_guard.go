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
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	// We use \b for word boundaries to avoid blocking things like "orders" (contains "order").
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME)\b`

	allowed := map[string]bool{
		"orders":             true,
		"order_line_items":   true,
		"inventory_items":    true,
		"inventory_mappings": true,
		"customers":          true,
		"customer_feedback":  true,
		"sources":            true,
		"feedback_statuses":  true,
		"planner_boards":     true,
		"planner_columns":    true,
		"planner_tasks":      true,
		"planner_sprints":    true,
	}

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		allowedTables:   allowed,
	}
}

// NormalizeTableName removes quotes and schema prefixes.
func (g *QueryGuard) NormalizeTableName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.Trim(name, "\"")
	if parts := strings.Split(name, "."); len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return name
}

// IsSafe checks if the SQL string contains any blocked mutation keywords and validates table references.
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

	// 2. Validate table references
	// We look for table identifiers following FROM or JOIN
	// This regex searches for the keyword and captures the FOLLOWING word(s) until a boundary.
	// To handle multiple tables, we look for all JOINs and commas.

	// First, let's find the FROM section
	fromIdx := regexp.MustCompile(`(?i)\bFROM\b`).FindStringIndex(normalized)
	if fromIdx == nil {
		// Might be SELECT 1+1
		return nil
	}

	// Capture everything after FROM
	afterFrom := normalized[fromIdx[1]:]

	// Truncate at keywords that end the table list
	endKeywords := []string{"WHERE", "GROUP", "ORDER", "LIMIT", "UNION", "SELECT"}
	for _, kw := range endKeywords {
		kwPattern := regexp.MustCompile(`(?i)\b` + kw + `\b`)
		if loc := kwPattern.FindStringIndex(afterFrom); loc != nil {
			afterFrom = afterFrom[:loc[0]]
		}
	}

	// Now we have a string like "orders o JOIN customers c ON ... JOIN line_items li ON ..."
	// or "orders, customers"

	// Split by JOIN keywords and commas to find potential table starts
	// We want to keep the "JOIN" parts to know we are looking at a table.
	// Actually, easier to just use a regex that finds words after FROM, JOIN, or comma.

	tableSeeker := regexp.MustCompile(`(?i)(?:\b(?:FROM|JOIN|INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|CROSS\s+JOIN)\b|,)\s+([a-z0-9_".]+|"[a-z0-9_.]+")`)

	// We need to prepend a comma or FROM to the afterFrom string so the regex catches the first table if it's not caught by the tableSeeker
	// But afterFrom already starts AFTER the first FROM.

	// Let's get the first table manually
	firstPart := strings.Fields(strings.Split(afterFrom, ",")[0])[0]
	if firstPart != "" {
		if err := g.validateTable(firstPart); err != nil {
			return err
		}
	}

	matches := tableSeeker.FindAllStringSubmatch(afterFrom, -1)
	for _, match := range matches {
		if len(match) > 1 {
			if err := g.validateTable(match[1]); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (g *QueryGuard) validateTable(name string) error {
	name = g.NormalizeTableName(name)
	// Ignore subqueries or common SQL keywords that might be caught
	if strings.HasPrefix(name, "(") || name == "" {
		return nil
	}
	// Some keywords might be caught if they follow a comma or JOIN incorrectly
	ignored := map[string]bool{"select": true, "on": true, "using": true, "as": true}
	if ignored[name] {
		return nil
	}

	if !g.allowedTables[name] {
		return fmt.Errorf("SECURITY ALERT: access to table '%s' is not allowed", name)
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
