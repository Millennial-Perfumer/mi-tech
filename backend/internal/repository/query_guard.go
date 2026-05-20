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

	allowed := []string{
		"orders", "order_line_items", "inventory_items", "inventory_mappings", "inventory_logs",
		"customers", "customer_feedback", "feedback_statuses", "oil_inventory", "suppliers",
		"purchase_orders", "manufacturing_records", "manufacturing_oils", "manufacturing_products",
		"planner_tasks", "planner_columns", "planner_boards", "planner_sprints", "planner_task_logs",
		"social_metrics_history", "social_post_history", "sources", "webhook_events",
	}

	allowedMap := make(map[string]bool)
	for _, t := range allowed {
		allowedMap[t] = true
	}

	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
		allowedTables:   allowedMap,
	}
}

// IsSafe checks if the SQL string contains any blocked mutation keywords or unauthorized tables.
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

	// Table Access Control: Verify all tables mentioned are in the allowlist.
	// We look for patterns like "FROM table_name" or "JOIN table_name".
	// This regex captures table names after FROM or JOIN, ignoring subqueries or aliases for simplicity.
	tableRegex := regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+([a-z0-9_]+)`)
	matches := tableRegex.FindAllStringSubmatch(normalized, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tableName := strings.ToLower(match[1])
			if !g.allowedTables[tableName] {
				return fmt.Errorf("SECURITY ALERT: unauthorized table access detected: %s", tableName)
			}
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
