package repository

import (
	"testing"
)

func TestQueryGuard_IsSafe(t *testing.T) {
	g := NewQueryGuard()

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "Simple valid SELECT",
			sql:     "SELECT * FROM orders",
			wantErr: false,
		},
		{
			name:    "Valid SELECT with JOIN",
			sql:     "SELECT o.id, c.name FROM orders o JOIN customers c ON o.customer_id = c.id",
			wantErr: false,
		},
		{
			name:    "Mutation keyword INSERT",
			sql:     "INSERT INTO orders (id) VALUES (1)",
			wantErr: true,
		},
		{
			name:    "Mutation keyword UPDATE",
			sql:     "UPDATE orders SET status = 'paid' WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Mutation keyword DELETE",
			sql:     "DELETE FROM orders",
			wantErr: true,
		},
		{
			name:    "Forbidden table users",
			sql:     "SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "Forbidden table app_configs",
			sql:     "SELECT * FROM app_configs",
			wantErr: true,
		},
		{
			name:    "Valid complex query",
			sql:     "SELECT title, sku, SUM(quantity) FROM order_line_items GROUP BY title, sku ORDER BY SUM(quantity) DESC LIMIT 10",
			wantErr: false,
		},
		{
			name:    "Schema prefix valid",
			sql:     "SELECT * FROM public.orders",
			wantErr: false,
		},
		{
			name:    "Quoted table name valid",
			sql:     "SELECT * FROM \"orders\"",
			wantErr: false,
		},
		{
			name:    "Multiple tables in FROM",
			sql:     "SELECT * FROM orders, customers WHERE orders.customer_id = customers.id",
			wantErr: false,
		},
		{
			name:    "Unauthorized table in JOIN",
			sql:     "SELECT * FROM orders JOIN users ON orders.user_id = users.id",
			wantErr: true,
		},
		{
			name:    "Case insensitive keyword block",
			sql:     "select * from orders; drop table orders",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := g.IsSafe(tt.sql); (err != nil) != tt.wantErr {
				t.Errorf("QueryGuard.IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
