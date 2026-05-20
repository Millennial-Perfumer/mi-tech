package repository

import (
	"testing"
)

func TestQueryGuard_IsSafe(t *testing.T) {
	guard := NewQueryGuard()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "Valid business query",
			query:   "SELECT * FROM orders LIMIT 10",
			wantErr: false,
		},
		{
			name:    "Valid join query",
			query:   "SELECT o.order_number, c.first_name FROM orders o JOIN customers c ON o.customer_phone = c.phone_number",
			wantErr: false,
		},
		{
			name:    "Mutation blocked (INSERT)",
			query:   "INSERT INTO orders (id) VALUES (1)",
			wantErr: true,
		},
		{
			name:    "Mutation blocked (UPDATE)",
			query:   "UPDATE orders SET status = 'paid' WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Sensitive table blocked (users)",
			query:   "SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "Sensitive table blocked (app_configs)",
			query:   "SELECT * FROM app_configs",
			wantErr: true,
		},
		{
			name:    "Sensitive table blocked in JOIN",
			query:   "SELECT * FROM orders JOIN users ON orders.id = users.id",
			wantErr: true,
		},
		{
			name:    "Unauthorized system table blocked",
			query:   "SELECT * FROM pg_users",
			wantErr: true,
		},
		{
			name:    "Non-SELECT query blocked",
			query:   "DROP TABLE orders",
			wantErr: true,
		},
		{
			name:    "Query with multiple allowed tables",
			query:   "SELECT * FROM inventory_items JOIN inventory_mappings ON inventory_items.mi_sku = inventory_mappings.mi_sku",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := guard.IsSafe(tt.query); (err != nil) != tt.wantErr {
				t.Errorf("QueryGuard.IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
