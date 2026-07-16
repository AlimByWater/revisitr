package botmanager

import "testing"

func TestValidateTableNum(t *testing.T) {
	valid := []struct{ in, want string }{
		{"7", "7"},
		{" 12 ", "12"},
		{"A5", "A5"},
		{"терраса", "терраса"},
	}
	for _, tc := range valid {
		got, err := validateTableNum(tc.in)
		if err != nil {
			t.Errorf("validateTableNum(%q): unexpected error %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("validateTableNum(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	invalid := []string{"", "   ", "123456789", "стол №7", "7-a", "7 8"}
	for _, in := range invalid {
		if _, err := validateTableNum(in); err == nil {
			t.Errorf("validateTableNum(%q): expected error", in)
		}
	}
}
