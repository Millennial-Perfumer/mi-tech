package repository

import (
	"testing"
)

func TestQueryGuard_TableAllowlist(t *testing.T) {
	g := NewQueryGuard()

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "Allowed table",
			sql:     "SELECT * FROM orders",
			wantErr: false,
		},
		{
			name:    "Blocked table - users",
			sql:     "SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "Blocked table - app_configs",
			sql:     "SELECT * FROM app_configs",
			wantErr: true,
		},
		{
			name:    "Complex query allowed",
			sql:     "SELECT o.id, li.title FROM orders o JOIN order_line_items li ON o.id = li.order_id",
			wantErr: false,
		},
		{
			name:    "Complex query blocked",
			sql:     "SELECT * FROM orders JOIN users ON orders.user_id = users.id",
			wantErr: true,
		},
		{
			name:    "Blocked table - users with aliases",
			sql:     "SELECT * FROM users u",
			wantErr: true,
		},
		{
			name:    "Subquery blocked",
			sql:     "SELECT * FROM (SELECT * FROM users) as t",
			wantErr: true,
		},
		{
			name:    "Comma separated tables allowed",
			sql:     "SELECT * FROM orders, customers",
			wantErr: false,
		},
		{
			name:    "Comma separated tables blocked",
			sql:     "SELECT * FROM orders, users",
			wantErr: true,
		},
		{
			name:    "Quoted schema and table allowed",
			sql:     `SELECT * FROM "public"."orders"`,
			wantErr: false,
		},
		{
			name:    "Quoted schema and table blocked",
			sql:     `SELECT * FROM "public"."users"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := g.IsSafe(tt.sql); (err != nil) != tt.wantErr {
				t.Errorf("IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
