package entity

import "testing"

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plus7 format", "+79991234567", "+79991234567"},
		{"8 format", "89991234567", "+79991234567"},
		{"7 without plus", "79991234567", "+79991234567"},
		{"10 digits", "9991234567", "+79991234567"},
		{"with dashes", "+7-999-123-45-67", "+79991234567"},
		{"with spaces", "+7 999 123 45 67", "+79991234567"},
		{"with parens", "+7(999)1234567", "+79991234567"},
		{"8 with formatting", "8 (999) 123-45-67", "+79991234567"},
		{"too short", "12345", ""},
		{"too long", "123456789012345", ""},
		{"empty", "", ""},
		{"letters only", "abcdef", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePhone(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
