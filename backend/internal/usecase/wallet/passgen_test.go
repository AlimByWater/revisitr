package wallet

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"revisitr/internal/entity"
)

// TestGeneratePass_Structure verifies that a generated .pkpass bundle is
// well-formed: it contains the entries Apple Wallet requires and the
// manifest.json SHA-1 hashes match the bundled files. Uses a self-signed
// test certificate in testdata/.
func TestGeneratePass_Structure(t *testing.T) {
	b64, err := os.ReadFile("testdata/test_cert.p12.b64")
	if err != nil {
		t.Fatal(err)
	}

	cfg := &entity.WalletConfig{
		Credentials: entity.WalletCredentials{
			"certificate":  string(bytes.TrimSpace(b64)),
			"pass_type_id": "pass.revisitr.test",
			"team_id":      "TEAM123456",
		},
		Design: entity.WalletDesign{
			OrganizationName: "Test Org",
			Description:      "Loyalty card",
			BackgroundColor:  "#ff0000",
			ForegroundColor:  "#ffffff",
		},
	}
	pass := &entity.WalletPass{
		SerialNumber: "abcd1234",
		AuthToken:    "tok",
		LastBalance:  250,
		LastLevel:    "Gold",
		ClientID:     10,
	}

	data, err := NewPassGenerator().GeneratePass(pass, cfg, "qr-value", "Test Org", "https://example.com/api/v1/wallet")
	if err != nil {
		t.Fatalf("GeneratePass: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("not a valid zip: %v", err)
	}

	contents := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(rc)
		rc.Close()
		contents[f.Name] = b
	}

	for _, required := range []string{"pass.json", "manifest.json", "signature", "icon.png"} {
		if _, ok := contents[required]; !ok {
			t.Errorf("missing required bundle entry: %s", required)
		}
	}

	// manifest.json hashes must match the bundled files.
	var manifest map[string]string
	if err := json.Unmarshal(contents["manifest.json"], &manifest); err != nil {
		t.Fatalf("manifest.json invalid: %v", err)
	}
	for name, want := range manifest {
		got := fmt.Sprintf("%x", sha1.Sum(contents[name]))
		if got != want {
			t.Errorf("manifest hash mismatch for %s: manifest=%s actual=%s", name, want, got)
		}
	}
	if manifest["pass.json"] == "" || manifest["icon.png"] == "" {
		t.Error("manifest must list pass.json and icon.png")
	}
}
