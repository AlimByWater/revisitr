package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func testP8PEM(t *testing.T) (string, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	return string(pemData), key
}

func TestProviderToken_ValidJWT(t *testing.T) {
	keyPEM, key := testP8PEM(t)
	a := NewAPNsClient()

	token, err := a.providerToken(keyPEM, "KEY123", "TEAM456")
	if err != nil {
		t.Fatalf("providerToken: %v", err)
	}

	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		return &key.PublicKey, nil
	}, jwt.WithValidMethods([]string{"ES256"}))
	if err != nil {
		t.Fatalf("parse signed jwt: %v", err)
	}
	if kid, _ := parsed.Header["kid"].(string); kid != "KEY123" {
		t.Errorf("kid = %q, want KEY123", kid)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["iss"] != "TEAM456" {
		t.Errorf("iss = %v, want TEAM456", claims["iss"])
	}
}

func TestProviderToken_Cached(t *testing.T) {
	keyPEM, _ := testP8PEM(t)
	a := NewAPNsClient()

	t1, err := a.providerToken(keyPEM, "KEY123", "TEAM456")
	if err != nil {
		t.Fatal(err)
	}
	t2, err := a.providerToken(keyPEM, "KEY123", "TEAM456")
	if err != nil {
		t.Fatal(err)
	}
	if t1 != t2 {
		t.Error("expected cached token to be reused")
	}
}

func TestParseP8Key_Base64(t *testing.T) {
	keyPEM, _ := testP8PEM(t)
	// base64-encoded PEM should also parse.
	if _, err := parseP8Key(base64.StdEncoding.EncodeToString([]byte(keyPEM))); err != nil {
		t.Fatalf("base64 PEM: %v", err)
	}
	// raw PEM should parse.
	if _, err := parseP8Key(keyPEM); err != nil {
		t.Fatalf("raw PEM: %v", err)
	}
	if _, err := parseP8Key("not a key"); err == nil {
		t.Error("expected error for garbage input")
	}
}
