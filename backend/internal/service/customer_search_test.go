package service

import (
	"reflect"
	"testing"
)

func TestCustomerService_ParseSearchQuery(t *testing.T) {
	s := &CustomerService{}

	tests := []struct {
		name   string
		search string
		want   CustomerFilter
	}{
		{
			name:   "empty fields",
			search: "first_name = '' email = ''",
			want: CustomerFilter{
				Search:         "",
				FirstNameEmpty: true,
				EmailEmpty:     true,
			},
		},
		{
			name:   "range fields",
			search: "spent > 1000 orders > 5",
			want: CustomerFilter{
				Search:    "",
				MinSpent:  1000,
				MinOrders: 5,
			},
		},
		{
			name:   "kv fields",
			search: "city:Mumbai state=MH email:test@example.com",
			want: CustomerFilter{
				Search: "",
				City:   "Mumbai",
				State:  "MH",
				Email:  "test@example.com",
			},
		},
		{
			name:   "mixed search",
			search: "John spent > 100 city:Mumbai",
			want: CustomerFilter{
				Search:   "John",
				MinSpent: 100,
				City:     "Mumbai",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.parseSearchQuery(tt.search); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CustomerService.parseSearchQuery() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
