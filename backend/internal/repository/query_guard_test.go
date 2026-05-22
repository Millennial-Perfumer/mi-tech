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
			name:    "Allowed SELECT",
			sql:     "SELECT * FROM orders",
			wantErr: false,
		},
		{
			name:    "Allowed JOIN",
			sql:     "SELECT * FROM orders JOIN customers ON orders.customer_phone = customers.phone_number",
			wantErr: false,
		},
		{
			name:    "Blocked INSERT",
			sql:     "INSERT INTO orders (id) VALUES (1)",
			wantErr: true,
		},
		{
			name:    "Blocked UPDATE",
			sql:     "UPDATE orders SET status = 'paid' WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Blocked DELETE",
			sql:     "DELETE FROM orders WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Blocked DROP",
			sql:     "DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "Blocked multiple keywords",
			sql:     "SELECT * FROM orders; DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "Case insensitive check",
			sql:     "select * from orders",
			wantErr: false,
		},
		{
			name:    "Keyword as part of word",
			sql:     "SELECT * FROM orders_history",
			wantErr: true,
		},
		{
			name:    "Blocked sensitive table users",
			sql:     "SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "Blocked sensitive table app_configs",
			sql:     "SELECT * FROM app_configs",
			wantErr: true,
		},
		{
			name:    "Blocked sensitive table app_settings",
			sql:     "SELECT * FROM app_settings",
			wantErr: true,
		},
		{
			name:    "Quoted table name",
			sql:     "SELECT * FROM \"orders\"",
			wantErr: false,
		},
		{
			name:    "Schema prefixed table",
			sql:     "SELECT * FROM public.orders",
			wantErr: false,
		},
		{
			name:    "Multiple JOINS with allowlist",
			sql:     "SELECT * FROM orders JOIN order_line_items ON orders.id = order_line_items.order_id JOIN customers ON orders.customer_phone = customers.phone_number",
			wantErr: false,
		},
		{
			name:    "JOIN with blocked table",
			sql:     "SELECT * FROM orders JOIN users ON orders.user_id = users.id",
			wantErr: true,
		},
		{
			name:    "Comma separated tables (Allowed)",
			sql:     "SELECT * FROM orders, customers WHERE orders.customer_phone = customers.phone_number",
			wantErr: false,
		},
		{
			name:    "Comma separated tables (Blocked second table)",
			sql:     "SELECT * FROM orders, users WHERE orders.user_id = users.id",
			wantErr: true,
		},
		{
			name:    "Query with WHERE clause",
			sql:     "SELECT * FROM orders WHERE status = 'paid'",
			wantErr: false,
		},
		{
			name:    "Table with alias",
			sql:     "SELECT o.id FROM orders o",
			wantErr: false,
		},
		{
			name:    "Subquery (ignored for now, base tables checked)",
			sql:     "SELECT * FROM (SELECT * FROM orders) as t",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := g.IsSafe(tt.sql); (err != nil) != tt.wantErr {
				t.Errorf("QueryGuard.IsSafe() [%s] error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
