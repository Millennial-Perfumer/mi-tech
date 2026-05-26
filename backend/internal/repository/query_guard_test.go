package repository

import (
	"testing"
)

func TestQueryGuard_TableAllowlist(t *testing.T) {
	g := NewQueryGuard()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"Allowed table", "SELECT * FROM orders", false},
		{"Restricted table", "SELECT * FROM users", true},
		{"Restricted table with sensitive config", "SELECT * FROM app_configs", true},
		{"Multiple tables allowed", "SELECT * FROM orders JOIN customers ON orders.customer_phone = customers.phone_number", false},
		{"One restricted table in join", "SELECT * FROM orders JOIN users ON orders.id = users.id", true},
		{"Schema prefix allowed", "SELECT * FROM public.orders", false},
		{"Schema prefix restricted", "SELECT * FROM public.users", true},
		{"Quoted identifier allowed", "SELECT * FROM \"orders\"", false},
		{"Quoted identifier restricted", "SELECT * FROM \"users\"", true},
		{"Case insensitive allowed", "select * from ORDERS", false},
		{"Multiple FROM allowed", "SELECT * FROM orders, customers", false},
		{"Multiple FROM mixed", "SELECT * FROM orders, users", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.IsSafe(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryGuard_MutationBlocking(t *testing.T) {
	g := NewQueryGuard()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"Valid SELECT", "SELECT * FROM orders", false},
		{"INSERT attempt", "INSERT INTO orders (id) VALUES (1)", true},
		{"UPDATE attempt", "UPDATE orders SET status = 'cancelled'", true},
		{"DELETE attempt", "DELETE FROM orders", true},
		{"DROP attempt", "DROP TABLE orders", true},
		{"UNION with restricted", "SELECT * FROM orders UNION SELECT * FROM users", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.IsSafe(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
