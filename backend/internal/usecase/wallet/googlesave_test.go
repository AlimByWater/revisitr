package wallet

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"

	"revisitr/internal/entity"
)

func TestGoogleSaveURL_NoIssuerID(t *testing.T) {
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{"service_account_key": "{}"}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "issuer_id") {
		t.Fatalf("expected issuer_id error, got %v", err)
	}
}

func TestGoogleSaveURL_NoServiceAccountKey(t *testing.T) {
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{"issuer_id": "test"}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "service_account_key") {
		t.Fatalf("expected service_account_key error, got %v", err)
	}
}

func TestGoogleSaveURL_InvalidServiceAccountJSON(t *testing.T) {
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{
			"issuer_id":          "test",
			"service_account_key": "not-json",
		}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "parse service account key") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestGoogleSaveURL_MissingClientEmail(t *testing.T) {
	sa, _ := json.Marshal(map[string]string{"private_key": "dGVzdA=="})
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{
			"issuer_id":          "test",
			"service_account_key": string(sa),
		}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "client_email") {
		t.Fatalf("expected client_email error, got %v", err)
	}
}

func TestGoogleSaveURL_MissingPrivateKey(t *testing.T) {
	sa, _ := json.Marshal(map[string]string{"client_email": "test@test.iam.gserviceaccount.com"})
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{
			"issuer_id":          "test",
			"service_account_key": string(sa),
		}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "private_key") {
		t.Fatalf("expected private_key error, got %v", err)
	}
}

func TestGoogleSaveURL_NoPEMBlock(t *testing.T) {
	sa, _ := json.Marshal(map[string]string{
		"client_email": "test@test.iam.gserviceaccount.com",
		"private_key":  "not-pem-data",
	})
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{
			"issuer_id":          "test",
			"service_account_key": string(sa),
		}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "no PEM block") {
		t.Fatalf("expected no PEM block error, got %v", err)
	}
}

func TestGoogleSaveURL_InvalidPEM(t *testing.T) {
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("not-a-valid-key"),
	})
	sa, _ := json.Marshal(map[string]string{
		"client_email": "test@test.iam.gserviceaccount.com",
		"private_key":  string(pemData),
	})
	gen := NewGoogleSaveGenerator()
	_, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1"},
		&entity.WalletConfig{Credentials: entity.WalletCredentials{
			"issuer_id":          "test",
			"service_account_key": string(sa),
		}},
		"", "",
	)
	if err == nil || !strings.Contains(err.Error(), "parse private key") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestGoogleSaveURL_SuccessMinimal(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})
	sa, _ := json.Marshal(map[string]string{
		"client_email": "test@test.iam.gserviceaccount.com",
		"private_key":  string(pemData),
	})
	gen := NewGoogleSaveGenerator()
	url, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1", ClientID: 1, LastBalance: 100, LastLevel: "Gold"},
		&entity.WalletConfig{
			OrgID: 1,
			Credentials: entity.WalletCredentials{
				"issuer_id":          "test-issuer",
				"service_account_key": string(sa),
			},
			Design: entity.WalletDesign{
				OrganizationName: "Test Org",
			},
		},
		"qr-value", "Test Org",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://pay.google.com/gp/v/save/") {
		t.Fatalf("unexpected url prefix: %s", url)
	}
}

func TestGoogleSaveURL_DefaultBackgroundColor(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})
	sa, _ := json.Marshal(map[string]string{
		"client_email": "test@test.iam.gserviceaccount.com",
		"private_key":  string(pemData),
	})
	gen := NewGoogleSaveGenerator()
	url, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1", ClientID: 1, LastBalance: 100, LastLevel: "Gold"},
		&entity.WalletConfig{
			OrgID: 1,
			Credentials: entity.WalletCredentials{
				"issuer_id":          "test-issuer",
				"service_account_key": string(sa),
			},
			Design: entity.WalletDesign{
				OrganizationName: "Test Org",
				BackgroundColor:  "",
			},
		},
		"qr-value", "Test Org",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://pay.google.com/gp/v/save/") {
		t.Fatalf("unexpected url prefix: %s", url)
	}
}

func TestGoogleSaveURL_WithAllDesignOptions(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})
	sa, _ := json.Marshal(map[string]string{
		"client_email": "test@test.iam.gserviceaccount.com",
		"private_key":  string(pemData),
	})
	gen := NewGoogleSaveGenerator()
	url, err := gen.GenerateSaveURL(
		&entity.WalletPass{SerialNumber: "s1", ClientID: 1, LastBalance: 100, LastLevel: "Gold"},
		&entity.WalletConfig{
			OrgID: 1,
			Credentials: entity.WalletCredentials{
				"issuer_id":          "test-issuer",
				"service_account_key": string(sa),
			},
			Design: entity.WalletDesign{
				OrganizationName: "Test Org",
				BackgroundColor:  "#ff0000",
				ForegroundColor:  "#ffffff",
				LabelColor:       "#cccccc",
				LogoURL:          "https://example.com/logo.png",
			},
		},
		"qr-value", "Test Org",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://pay.google.com/gp/v/save/") {
		t.Fatalf("unexpected url prefix: %s", url)
	}
}

func TestNewGoogleSaveGenerator(t *testing.T) {
	gen := NewGoogleSaveGenerator()
	if gen == nil {
		t.Fatal("expected non-nil generator")
	}
}
