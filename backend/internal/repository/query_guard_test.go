package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryGuard_IsSafe(t *testing.T) {
	guard := NewQueryGuard()

	tests := []struct {
		name     string
		query    string
		isSafe   bool
		errMatch string
	}{
		{
			name:   "Valid SELECT",
			query:  "SELECT * FROM orders",
			isSafe: true,
		},
		{
			name:   "Valid JOIN",
			query:  "SELECT * FROM orders JOIN order_line_items ON orders.id = order_line_items.order_id",
			isSafe: true,
		},
		{
			name:   "Valid Multi-FROM",
			query:  "SELECT * FROM orders, order_line_items WHERE orders.id = order_line_items.order_id",
			isSafe: true,
		},
		{
			name:   "Valid Schema Prefix",
			query:  "SELECT * FROM public.orders",
			isSafe: true,
		},
		{
			name:   "Valid Quoted",
			query:  `SELECT * FROM "orders"`,
			isSafe: true,
		},
		{
			name:     "Mutation keyword: INSERT",
			query:    "INSERT INTO orders (id) VALUES (1)",
			isSafe:   false,
			errMatch: "only SELECT queries are allowed",
		},
		{
			name:     "Mutation keyword: embedded",
			query:    "SELECT * FROM orders; DROP TABLE users",
			isSafe:   false,
			errMatch: "mutation keyword detected",
		},
		{
			name:     "Sensitive table: users",
			query:    "SELECT * FROM users",
			isSafe:   false,
			errMatch: "access to table 'users' is not allowed",
		},
		{
			name:     "Sensitive table: app_configs",
			query:    "SELECT * FROM app_configs",
			isSafe:   false,
			errMatch: "access to table 'app_configs' is not allowed",
		},
		{
			name:     "Sensitive table in JOIN",
			query:    "SELECT * FROM orders JOIN users ON orders.customer_id = users.id",
			isSafe:   false,
			errMatch: "access to table 'users' is not allowed",
		},
		{
			name:     "Sensitive table in Multi-FROM",
			query:    "SELECT * FROM orders, users",
			isSafe:   false,
			errMatch: "access to table 'users' is not allowed",
		},
		{
			name:     "Sensitive table with Quoted and Prefix",
			query:    `SELECT * FROM public."users"`,
			isSafe:   false,
			errMatch: "access to table 'users' is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guard.IsSafe(tt.query)
			if tt.isSafe {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "SECURITY ALERT")
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
			}
		})
	}
}
