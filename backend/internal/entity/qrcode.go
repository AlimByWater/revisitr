package entity

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateQRCode creates a unique identifier for client QR code.
// Format: "RVS-XXXXXXXXXXXXXXXX" (prefix + 16 hex chars).
func GenerateQRCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("GenerateQRCode: %w", err)
	}
	return "RVS-" + hex.EncodeToString(b), nil
}
