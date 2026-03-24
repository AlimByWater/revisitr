package entity

import (
	"strings"
	"testing"
)

func TestGenerateQRCode(t *testing.T) {
	code, err := GenerateQRCode()
	if err != nil {
		t.Fatalf("GenerateQRCode() error: %v", err)
	}
	if !strings.HasPrefix(code, "RVS-") {
		t.Errorf("code %q does not have RVS- prefix", code)
	}
	// "RVS-" (4 chars) + 16 hex chars = 20 total
	if len(code) != 20 {
		t.Errorf("code length = %d, want 20", len(code))
	}
}

func TestGenerateQRCode_Unique(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := GenerateQRCode()
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if codes[code] {
			t.Fatalf("duplicate code at iteration %d: %s", i, code)
		}
		codes[code] = true
	}
}
