package wallet

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2/google"

	"revisitr/internal/entity"
)

type GoogleWalletAPI struct {
	logger *slog.Logger
	client *http.Client
}

func NewGoogleWalletAPI() *GoogleWalletAPI {
	return &GoogleWalletAPI{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (api *GoogleWalletAPI) Init(logger *slog.Logger) {
	api.logger = logger
}

// EnsureClass creates or updates the LoyaltyClass via REST API.
// The class ID is deterministic: {issuerID}.revisitr-class-{orgID}.
func (api *GoogleWalletAPI) EnsureClass(ctx context.Context, creds map[string]string, orgID int, orgName string, design entity.WalletDesign) error {
	issuerID := creds["issuer_id"]
	saKey := creds["service_account_key"]
	if issuerID == "" || saKey == "" {
		return fmt.Errorf("google wallet: missing issuer_id or service_account_key")
	}

	token, err := api.getToken(ctx, saKey)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	classID := fmt.Sprintf("%s.revisitr-class-%d", issuerID, orgID)

	bgColor := design.BackgroundColor
	if bgColor == "" {
		bgColor = "#1a1a2e"
	}

	body := map[string]any{
		"id":          classID,
		"issuerName":  orgName,
		"programName": orgName,
		"programLogo": map[string]any{
			"sourceUri": map[string]any{"uri": api.logoURL(design)},
		},
		"reviewStatus":       "underReview",
		"hexBackgroundColor": bgColor,
	}
	if design.ForegroundColor != "" {
		body["hexForegroundColor"] = design.ForegroundColor
	}
	if design.LabelColor != "" {
		body["hexLabelColor"] = design.LabelColor
	}

	resp := api.restCall(ctx, "POST",
		"https://walletobjects.googleapis.com/walletobjects/v1/loyaltyClass",
		token, body)
	if resp == nil {
		// Class may already exist — update it
		resp = api.restCall(ctx, "PUT",
			fmt.Sprintf("https://walletobjects.googleapis.com/walletobjects/v1/loyaltyClass/%s", classID),
			token, body)
	}
	if resp == nil {
		return fmt.Errorf("google wallet: failed to create/update class")
	}
	return nil
}

// CreateObject creates a LoyaltyObject via REST API.
// Object ID: {issuerID}.{serialNumber}.
func (api *GoogleWalletAPI) CreateObject(ctx context.Context, creds map[string]string, orgID int, pass *entity.WalletPass, clientQR, orgName string, design entity.WalletDesign) error {
	issuerID := creds["issuer_id"]
	saKey := creds["service_account_key"]
	if issuerID == "" || saKey == "" {
		return fmt.Errorf("google wallet: missing issuer_id or service_account_key")
	}

	token, err := api.getToken(ctx, saKey)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	classID := fmt.Sprintf("%s.revisitr-class-%d", issuerID, orgID)
	objID := fmt.Sprintf("%s.%s", issuerID, pass.SerialNumber)

	bgColor := design.BackgroundColor
	if bgColor == "" {
		bgColor = "#1a1a2e"
	}

	body := map[string]any{
		"id":      objID,
		"classId": classID,
		"state":   "ACTIVE",
		"accountId":   fmt.Sprintf("client_%d", pass.ClientID),
		"accountName": orgName,
		"loyaltyPoints": map[string]any{
			"label":   "Баллы",
			"balance": map[string]any{"string": fmt.Sprintf("%d", pass.LastBalance)},
		},
		"barcode": map[string]any{
			"type":  "QR_CODE",
			"value": clientQR,
		},
		"hexBackgroundColor": bgColor,
		"cardTitle": map[string]any{
			"defaultValue": map[string]any{"language": "en-US", "value": orgName},
		},
		"header": map[string]any{
			"defaultValue": map[string]any{"language": "en-US", "value": fmt.Sprintf("%d баллов", pass.LastBalance)},
		},
	}
	if design.ForegroundColor != "" {
		body["hexForegroundColor"] = design.ForegroundColor
	}
	if design.LabelColor != "" {
		body["hexLabelColor"] = design.LabelColor
	}
	if pass.LastLevel != "" {
		body["secondary"] = []map[string]any{
			{
				"header": "Уровень",
				"value":  map[string]any{"defaultValue": map[string]any{"language": "en-US", "value": pass.LastLevel}},
			},
		}
	}

	resp := api.restCall(ctx, "POST",
		"https://walletobjects.googleapis.com/walletobjects/v1/loyaltyObject",
		token, body)
	if resp == nil {
		// Object may already exist — update it
		resp = api.restCall(ctx, "PUT",
			fmt.Sprintf("https://walletobjects.googleapis.com/walletobjects/v1/loyaltyObject/%s", objID),
			token, body)
	}
	if resp == nil {
		return fmt.Errorf("google wallet: failed to create object")
	}
	return nil
}

// GenerateSaveURL creates a signed JWT referencing an existing LoyaltyObject.
// This is used for the "Add to Google Wallet" button.
func (api *GoogleWalletAPI) GenerateSaveURL(creds map[string]string, orgID int, pass *entity.WalletPass) (string, error) {
	issuerID := creds["issuer_id"]
	saKey := creds["service_account_key"]
	if issuerID == "" || saKey == "" {
		return "", fmt.Errorf("google wallet: missing issuer_id or service_account_key")
	}

	classID := fmt.Sprintf("%s.revisitr-class-%d", issuerID, orgID)
	objID := fmt.Sprintf("%s.%s", issuerID, pass.SerialNumber)

	var sa struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(saKey), &sa); err != nil {
		return "", fmt.Errorf("google wallet: parse service account key: %w", err)
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return "", fmt.Errorf("google wallet: service_account_key missing client_email or private_key")
	}

	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("google wallet: no PEM block in private_key")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("google wallet: parse private key: %w", err)
		}
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("google wallet: private key is not RSA")
	}

	payload := map[string]any{
		"loyaltyObjects": []any{
			map[string]any{"id": objID, "classId": classID},
		},
	}

	claims := jwt.MapClaims{
		"iss":     sa.ClientEmail,
		"aud":     "google",
		"origins": []string{},
		"typ":     "savetowallet",
		"iat":     time.Now().Unix(),
		"payload": payload,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(rsaKey)
	if err != nil {
		return "", fmt.Errorf("google wallet: sign jwt: %w", err)
	}

	return "https://pay.google.com/gp/v/save/" + signed, nil
}

func (api *GoogleWalletAPI) getToken(ctx context.Context, saKey string) (string, error) {
	conf, err := google.JWTConfigFromJSON([]byte(saKey), "https://www.googleapis.com/auth/wallet_object.issuer")
	if err != nil {
		return "", fmt.Errorf("parse jwt config: %w", err)
	}
	token, err := conf.TokenSource(ctx).Token()
	if err != nil {
		return "", fmt.Errorf("get oauth2 token: %w", err)
	}
	return token.AccessToken, nil
}

func (api *GoogleWalletAPI) restCall(ctx context.Context, method, url, accessToken string, body any) map[string]any {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			if api.logger != nil {
				api.logger.Error("google wallet: marshal request", "error", err)
			}
			return nil
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		if api.logger != nil {
			api.logger.Error("google wallet: create request", "error", err)
		}
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		if api.logger != nil {
			api.logger.Error("google wallet: api call", "error", err, "method", method, "url", url)
		}
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return map[string]any{"ok": true}
		}
		return result
	}

	respBody, _ := io.ReadAll(resp.Body)
	if api.logger != nil {
		api.logger.Warn("google wallet: api error",
			"status", resp.StatusCode,
			"body", truncate(string(respBody), 500),
			"method", method,
			"url", url,
		)
	}
	return nil
}

func (api *GoogleWalletAPI) logoURL(design entity.WalletDesign) string {
	if design.LogoURL != "" {
		return design.LogoURL
	}
	return "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
