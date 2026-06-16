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
}

func NewQueryGuard() *QueryGuard {
	// Pattern blocks common mutation keywords. Case-insensitive.
	// We use \b for word boundaries to avoid blocking things like "orders" (contains "order").
	// Removed REPLACE because it's a common SELECT function.
	// Removed COMMENT because it triggers on SQL comments.
	pattern := `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|ATTACH|DETACH|MERGE|UPSERT|RENAME|COPY|DO|CALL|EXECUTE)\b`
	return &QueryGuard{
		blockedPatterns: regexp.MustCompile(pattern),
	}
}

// IsSafe checks if the SQL string contains any blocked mutation keywords.
func (g *QueryGuard) IsSafe(sql string) error {
	// Stacked query prevention
	// Remove trailing semicolons and whitespace first
	cleaned := strings.TrimRight(strings.TrimSpace(sql), "; \t\n\r")

	// A naive check: if there is another semicolon, it might be a stacked query.
	// To avoid breaking legitimate queries with semicolons in string literals,
	// we do a simple string literal strip.
	inString := false
	for _, c := range cleaned {
		if c == '\'' {
			inString = !inString
		}
		if c == ';' && !inString {
			return fmt.Errorf("SECURITY ALERT: stacked queries (;) are not allowed")
		}
	}

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

	return nil
}

// WrapQuery is a helper to validate a query before returning it.
func (g *QueryGuard) WrapQuery(sql string) (string, error) {
	if err := g.IsSafe(sql); err != nil {
		return "", err
	}
	return sql, nil
}
