package repository

import (
	"strings"
	"testing"
)

func TestQueryGuard_IsSafe(t *testing.T) {
	guard := NewQueryGuard()

	tests := []struct {
		name    string
		query   string
		isSafe  bool
		wantErr string
	}{
		{
			name:   "Safe SELECT",
			query:  "SELECT * FROM orders",
			isSafe: true,
		},
		{
			name:   "Safe SELECT with JOIN",
			query:  "SELECT * FROM orders JOIN order_line_items ON orders.id = order_line_items.order_id",
			isSafe: true,
		},
		{
			name:   "Safe SELECT with quotes",
			query:  "SELECT * FROM \"orders\"",
			isSafe: true,
		},
		{
			name:   "Safe SELECT with schema",
			query:  "SELECT * FROM public.orders",
			isSafe: true,
		},
		{
			name:    "Blocked INSERT",
			query:   "INSERT INTO orders (id) VALUES (1)",
			isSafe:  false,
			wantErr: "only SELECT queries are allowed",
		},
		{
			name:    "Blocked mutation keyword in SELECT",
			query:   "SELECT * FROM orders; DROP TABLE orders",
			isSafe:  false,
			wantErr: "mutation keyword detected",
		},
		{
			name:    "Unauthorized table access",
			query:   "SELECT * FROM users",
			isSafe:  false,
			wantErr: "unauthorized table access detected: users",
		},
		{
			name:    "Unauthorized table access in JOIN",
			query:   "SELECT * FROM orders JOIN users ON orders.customer_email = users.username",
			isSafe:  false,
			wantErr: "unauthorized table access detected: users",
		},
		{
			name:    "Unauthorized table access with schema",
			query:   "SELECT * FROM information_schema.tables",
			isSafe:  false,
			wantErr: "unauthorized table access detected: tables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guard.IsSafe(tt.query)
			if (err == nil) != tt.isSafe {
				t.Errorf("IsSafe() error = %v, wantSafe %v", err, tt.isSafe)
			}
			if err != nil && tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
