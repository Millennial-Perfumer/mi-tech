package repository

import (
	"testing"
)

func TestQueryGuard_IsSafe(t *testing.T) {
	guard := NewQueryGuard()

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "Valid simple select",
			sql:     "SELECT * FROM orders",
			wantErr: false,
		},
		{
			name:    "Valid select with join",
			sql:     "SELECT * FROM orders JOIN order_line_items ON orders.id = order_line_items.order_id",
			wantErr: false,
		},
		{
			name:    "Valid select with where",
			sql:     "SELECT name FROM customers WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "Blocked mutation: INSERT",
			sql:     "INSERT INTO orders (id) VALUES ('1')",
			wantErr: true,
		},
		{
			name:    "Blocked mutation: DROP",
			sql:     "DROP TABLE orders",
			wantErr: true,
		},
		{
			name:    "Restricted table: users",
			sql:     "SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "Restricted table: app_configs",
			sql:     "SELECT * FROM app_configs",
			wantErr: true,
		},
		{
			name:    "Restricted table in JOIN",
			sql:     "SELECT * FROM orders JOIN users ON orders.customer_name = users.username",
			wantErr: true,
		},
		{
			name:    "Case insensitivity check",
			sql:     "select * from ORDERS",
			wantErr: false,
		},
		{
			name:    "Schema prefix handling",
			sql:     "SELECT * FROM public.orders",
			wantErr: false,
		},
		{
			name:    "Unauthorized table with schema prefix",
			sql:     "SELECT * FROM public.users",
			wantErr: true,
		},
		{
			name:    "Non-SELECT query",
			sql:     "WITH t AS (SELECT 1) SELECT * FROM t",
			wantErr: true,
		},
		{
			name:    "Comma separated tables: Valid",
			sql:     "SELECT * FROM orders, order_line_items WHERE orders.id = order_line_items.order_id",
			wantErr: false,
		},
		{
			name:    "Comma separated tables: One Invalid",
			sql:     "SELECT * FROM orders, users",
			wantErr: true,
		},
		{
			name:    "Table aliases: Valid",
			sql:     "SELECT o.id FROM orders o WHERE o.total_price > 100",
			wantErr: false,
		},
		{
			name:    "Table aliases: Invalid",
			sql:     "SELECT u.id FROM users u",
			wantErr: true,
		},
		{
			name:    "Subquery: Restricted",
			sql:     "SELECT * FROM (SELECT * FROM orders) as t",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := guard.IsSafe(tt.sql); (err != nil) != tt.wantErr {
				t.Errorf("QueryGuard.IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryGuard_IsTableAllowed(t *testing.T) {
	guard := NewQueryGuard()

	tests := []struct {
		tableName string
		want      bool
	}{
		{"orders", true},
		{"ORDERS", true},
		{"users", false},
		{"app_configs", false},
		{"public.orders", true},
		{"public.users", false},
	}

	for _, tt := range tests {
		t.Run(tt.tableName, func(t *testing.T) {
			if got := guard.IsTableAllowed(tt.tableName); got != tt.want {
				t.Errorf("QueryGuard.IsTableAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
