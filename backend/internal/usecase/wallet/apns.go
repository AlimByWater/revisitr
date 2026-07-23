package wallet

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// apnsHost is the production APNs endpoint. Wallet pass updates always use
// production APNs regardless of the signing certificate.
const apnsHost = "https://api.push.apple.com"

// APNsClient sends Apple Wallet update pushes using token-based (.p8) auth.
type APNsClient struct {
	logger *slog.Logger
	client *http.Client

	mu     sync.Mutex
	tokens map[string]cachedToken // keyed by APNs key ID
}

type cachedToken struct {
	jwt       string
	issuedAt  time.Time
}

func NewAPNsClient() *APNsClient {
	return &APNsClient{
		client: &http.Client{Timeout: 15 * time.Second},
		tokens: map[string]cachedToken{},
	}
}

func (a *APNsClient) Init(logger *slog.Logger) {
	a.logger = logger
}

// SendPush notifies a device that its pass changed. The payload is empty per the
// PassKit protocol — the device then fetches the pass via the web service.
// creds must contain: apns_key (base64 or PEM .p8), apns_key_id, team_id, pass_type_id.
func (a *APNsClient) SendPush(ctx context.Context, creds map[string]string, pushToken string) error {
	token, err := a.providerToken(creds["apns_key"], creds["apns_key_id"], creds["team_id"])
	if err != nil {
		return fmt.Errorf("apns provider token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apnsHost+"/3/device/"+pushToken, bytes.NewReader([]byte("{}")))
	if err != nil {
		return fmt.Errorf("apns new request: %w", err)
	}
	req.Header.Set("authorization", "bearer "+token)
	req.Header.Set("apns-topic", creds["pass_type_id"])
	req.Header.Set("content-type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("apns request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apns status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// providerToken returns a cached APNs JWT, regenerating it when older than ~50
// minutes (APNs rejects tokens older than 1 hour and rate-limits regeneration).
func (a *APNsClient) providerToken(keyPEM, keyID, teamID string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if t, ok := a.tokens[keyID]; ok && time.Since(t.issuedAt) < 50*time.Minute {
		return t.jwt, nil
	}

	key, err := parseP8Key(keyPEM)
	if err != nil {
		return "", err
	}

	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
	})
	tok.Header["kid"] = keyID

	signed, err := tok.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign apns jwt: %w", err)
	}

	a.tokens[keyID] = cachedToken{jwt: signed, issuedAt: now}
	return signed, nil
}

// parseP8Key parses an APNs .p8 private key, accepting either raw PEM text or a
// base64-encoded PEM/DER blob.
func parseP8Key(data string) (*ecdsa.PrivateKey, error) {
	raw := []byte(data)
	block, _ := pem.Decode(raw)
	if block == nil {
		// Not PEM — try base64 (the whole file base64-encoded).
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("apns key is neither PEM nor base64: %w", err)
		}
		if block, _ = pem.Decode(decoded); block == nil {
			// base64 of raw DER
			return ecKeyFromPKCS8(decoded)
		}
	}
	return ecKeyFromPKCS8(block.Bytes)
}

func ecKeyFromPKCS8(der []byte) (*ecdsa.PrivateKey, error) {
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("parse pkcs8 apns key: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("apns key is not an ECDSA key")
	}
	return key, nil
}
