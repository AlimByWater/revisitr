package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// WalletCredentials stores platform-specific credentials (JSONB).
// Apple: pass_type_id, team_id, certificate (base64-encoded .p12)
// Google: issuer_id, service_account_key (JSON string)
type WalletCredentials map[string]string

func (c WalletCredentials) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("WalletCredentials.Value: %w", err)
	}
	return b, nil
}

func (c *WalletCredentials) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = WalletCredentials{}
		return nil
	default:
		return fmt.Errorf("WalletCredentials.Scan: unsupported type %T", src)
	}
}

// WalletDesign stores pass appearance settings (JSONB).
type WalletDesign struct {
	LogoURL         string `json:"logo_url,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	ForegroundColor string `json:"foreground_color,omitempty"`
	LabelColor      string `json:"label_color,omitempty"`
	Description     string `json:"description,omitempty"`
}

func (d WalletDesign) Value() (driver.Value, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("WalletDesign.Value: %w", err)
	}
	return b, nil
}

func (d *WalletDesign) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	case string:
		return json.Unmarshal([]byte(v), d)
	case nil:
		*d = WalletDesign{}
		return nil
	default:
		return fmt.Errorf("WalletDesign.Scan: unsupported type %T", src)
	}
}

// WalletConfig holds per-org wallet platform configuration.
type WalletConfig struct {
	ID          int               `db:"id" json:"id"`
	OrgID       int               `db:"org_id" json:"org_id"`
	Platform    string            `db:"platform" json:"platform"`
	IsEnabled   bool              `db:"is_enabled" json:"is_enabled"`
	Credentials WalletCredentials `db:"credentials" json:"-"`
	Design      WalletDesign      `db:"design" json:"design"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
}

// WalletPass represents a single wallet pass issued to a client.
type WalletPass struct {
	ID            int        `db:"id" json:"id"`
	OrgID         int        `db:"org_id" json:"org_id"`
	ClientID      int        `db:"client_id" json:"client_id"`
	Platform      string     `db:"platform" json:"platform"`
	SerialNumber  string     `db:"serial_number" json:"serial_number"`
	AuthToken     string     `db:"auth_token" json:"-"`
	PushToken     *string    `db:"push_token" json:"-"`
	LastBalance   int        `db:"last_balance" json:"last_balance"`
	LastLevel     string     `db:"last_level" json:"last_level"`
	LastUpdatedAt *time.Time `db:"last_updated_at" json:"last_updated_at,omitempty"`
	Status        string     `db:"status" json:"status"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// --- Request / Response DTOs ---

type SaveWalletConfigRequest struct {
	Platform    string            `json:"platform" binding:"required,oneof=apple google"`
	IsEnabled   bool              `json:"is_enabled"`
	Credentials WalletCredentials `json:"credentials"`
	Design      WalletDesign      `json:"design"`
}

type IssueWalletPassRequest struct {
	ClientID int    `json:"client_id" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=apple google"`
}

type RegisterPushTokenRequest struct {
	PushToken string `json:"push_token" binding:"required"`
}

type WalletPassPublic struct {
	ID           int        `json:"id"`
	ClientID     int        `json:"client_id"`
	Platform     string     `json:"platform"`
	SerialNumber string     `json:"serial_number"`
	LastBalance  int        `json:"last_balance"`
	LastLevel    string     `json:"last_level"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

type WalletStats struct {
	TotalPasses  int `json:"total_passes"`
	ApplePasses  int `json:"apple_passes"`
	GooglePasses int `json:"google_passes"`
	ActivePasses int `json:"active_passes"`
}
