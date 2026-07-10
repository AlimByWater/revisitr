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

func TestAPI_GenerateSaveURL_MissingIssuerID(t *testing.T) {
	api := NewGoogleWalletAPI()
	_, err := api.GenerateSaveURL(
		map[string]string{"service_account_key": "{}"},
		1, &entity.WalletPass{SerialNumber: "s1"},
	)
	if err == nil || !strings.Contains(err.Error(), "issuer_id") {
		t.Fatalf("expected issuer_id error, got %v", err)
	}
}

func TestAPI_GenerateSaveURL_MissingServiceAccountKey(t *testing.T) {
	api := NewGoogleWalletAPI()
	_, err := api.GenerateSaveURL(
		map[string]string{"issuer_id": "test"},
		1, &entity.WalletPass{SerialNumber: "s1"},
	)
	if err == nil || !strings.Contains(err.Error(), "service_account_key") {
		t.Fatalf("expected service_account_key error, got %v", err)
	}
}

func TestAPI_GenerateSaveURL_InvalidJSON(t *testing.T) {
	api := NewGoogleWalletAPI()
	_, err := api.GenerateSaveURL(
		map[string]string{
			"issuer_id":          "test",
			"service_account_key": "not-json",
		},
		1, &entity.WalletPass{SerialNumber: "s1"},
	)
	if err == nil || !strings.Contains(err.Error(), "parse service account key") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestAPI_GenerateSaveURL_Success(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})
	saKey, _ := json.Marshal(map[string]string{
		"client_email": "test@test.iam.gserviceaccount.com",
		"private_key":  string(pemData),
	})

	api := NewGoogleWalletAPI()
	url, err := api.GenerateSaveURL(
		map[string]string{
			"issuer_id":          "test-issuer",
			"service_account_key": string(saKey),
		},
		42, &entity.WalletPass{SerialNumber: "abcd1234", ClientID: 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://pay.google.com/gp/v/save/") {
		t.Fatalf("unexpected url prefix: %s", url)
	}
}
