package wallet

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"revisitr/internal/entity"
)

type GoogleSaveGenerator struct{}

func NewGoogleSaveGenerator() *GoogleSaveGenerator {
	return &GoogleSaveGenerator{}
}

func (g *GoogleSaveGenerator) GenerateSaveURL(pass *entity.WalletPass, cfg *entity.WalletConfig, clientQR string, orgName string) (string, error) {
	issuerID := cfg.Credentials["issuer_id"]
	serviceAccountKey := cfg.Credentials["service_account_key"]
	if issuerID == "" || serviceAccountKey == "" {
		return "", errors.New("google wallet: issuer_id and service_account_key are required")
	}

	var sa struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(serviceAccountKey), &sa); err != nil {
		return "", fmt.Errorf("google wallet: parse service account key: %w", err)
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return "", errors.New("google wallet: service_account_key missing client_email or private_key")
	}

	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return "", errors.New("google wallet: no PEM block in private_key")
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
		return "", errors.New("google wallet: private key is not RSA")
	}

	classSuffix := fmt.Sprintf("revisitr-class-%d", cfg.OrgID)
	objectSuffix := pass.SerialNumber

	loyaltyClass := map[string]any{
		"id":         fmt.Sprintf("%s.%s", issuerID, classSuffix),
		"issuerName": orgName,
	}

	design := cfg.Design

	bgColor := design.BackgroundColor
	if bgColor == "" {
		bgColor = "#1a1a2e"
	}

	loyaltyObject := map[string]any{
		"id":      fmt.Sprintf("%s.%s", issuerID, objectSuffix),
		"classId": fmt.Sprintf("%s.%s", issuerID, classSuffix),
		"state":   "ACTIVE",
		"accountId": fmt.Sprintf("client_%d", pass.ClientID),
		"accountName": orgName,
		"loyaltyPoints": []map[string]any{
			{
				"label":   "Баллы",
				"balance": map[string]any{
					"string": fmt.Sprintf("%d", pass.LastBalance),
				},
			},
		},
		"barcode": map[string]any{
			"type":  "QR_CODE",
			"value": clientQR,
		},
		"hexBackgroundColor": bgColor,
		"cardTitle": map[string]any{
			"defaultValue": map[string]any{
				"language": "en-US",
				"value":    orgName,
			},
		},
		"header": map[string]any{
			"defaultValue": map[string]any{
				"language": "en-US",
				"value":    fmt.Sprintf("%d баллов", pass.LastBalance),
			},
		},
		"rewardProgramLabel": fmt.Sprintf("%d баллов", pass.LastBalance),
		"secondary": []map[string]any{
			{
				"header": "Уровень",
				"value": map[string]any{
					"defaultValue": map[string]any{
						"language": "en-US",
						"value":    pass.LastLevel,
					},
				},
			},
		},
	}

	if design.ForegroundColor != "" {
		loyaltyObject["hexForegroundColor"] = design.ForegroundColor
	}
	if design.LabelColor != "" {
		loyaltyObject["hexLabelColor"] = design.LabelColor
	}
	if design.LogoURL != "" {
		loyaltyObject["logo"] = map[string]any{
			"sourceUri": map[string]any{
				"uri": design.LogoURL,
			},
			"contentDescription": map[string]any{
				"defaultValue": map[string]any{
					"language": "en-US",
					"value":    "Logo",
				},
			},
		}
	}

	payload := map[string]any{
		"loyaltyClasses": []any{loyaltyClass},
		"loyaltyObjects": []any{loyaltyObject},
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