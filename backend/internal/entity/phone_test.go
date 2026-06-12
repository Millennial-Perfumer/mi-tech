package entity

import (
	"testing"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"9876543210", "+919876543210"},
		{"919876543210", "+919876543210"},
		{"+919876543210", "+919876543210"},
		{"  9876543210  ", "+919876543210"},
		{"987-654-3210", "+919876543210"},
		{"+1 202-456-1111", "+12024561111"},
		{"+44 7911 123456", "+447911123456"},
		{"invalid phone", ""},
		{"+910000000000", "+910000000000"}, // Dummy number
	}

	for _, tc := range tests {
		actual := NormalizePhone(tc.input)
		if actual != tc.expected {
			t.Errorf("expected %s, got %s for input %s", tc.expected, actual, tc.input)
		}
	}
}
