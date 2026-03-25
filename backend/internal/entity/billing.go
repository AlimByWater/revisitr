package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// TariffFeatures controls which modules are available on a plan.
type TariffFeatures struct {
	Loyalty           bool `json:"loyalty"`
	Campaigns         bool `json:"campaigns"`
	Promotions        bool `json:"promotions"`
	Integrations      bool `json:"integrations"`
	Analytics         bool `json:"analytics"`
	RFM               bool `json:"rfm"`
	AdvancedCampaigns bool `json:"advanced_campaigns"`
}

func (f TariffFeatures) Value() (driver.Value, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("TariffFeatures.Value: %w", err)
	}
	return b, nil
}

func (f *TariffFeatures) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, f)
	case string:
		return json.Unmarshal([]byte(v), f)
	case nil:
		*f = TariffFeatures{}
		return nil
	default:
		return fmt.Errorf("TariffFeatures.Scan: unsupported type %T", src)
	}
}

// TariffLimits defines numeric limits per plan. -1 means unlimited.
type TariffLimits struct {
	MaxClients           int `json:"max_clients"`
	MaxBots              int `json:"max_bots"`
	MaxCampaignsPerMonth int `json:"max_campaigns_per_month"`
	MaxPOS               int `json:"max_pos"`
}

func (l TariffLimits) Value() (driver.Value, error) {
	b, err := json.Marshal(l)
	if err != nil {
		return nil, fmt.Errorf("TariffLimits.Value: %w", err)
	}
	return b, nil
}

func (l *TariffLimits) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, l)
	case string:
		return json.Unmarshal([]byte(v), l)
	case nil:
		*l = TariffLimits{}
		return nil
	default:
		return fmt.Errorf("TariffLimits.Scan: unsupported type %T", src)
	}
}

type Tariff struct {
	ID        int            `json:"id"         db:"id"`
	Name      string         `json:"name"       db:"name"`
	Slug      string         `json:"slug"       db:"slug"`
	Price     int            `json:"price"      db:"price"`         // kopeks
	Currency  string         `json:"currency"   db:"currency"`
	Interval  string         `json:"interval"   db:"interval"`      // "month" | "year"
	Features  TariffFeatures `json:"features"   db:"features"`
	Limits    TariffLimits   `json:"limits"     db:"limits"`
	Active    bool           `json:"active"     db:"active"`
	SortOrder int            `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

type Subscription struct {
	ID                 int        `json:"id"                   db:"id"`
	OrgID              int        `json:"org_id"               db:"org_id"`
	TariffID           int        `json:"tariff_id"            db:"tariff_id"`
	Status             string     `json:"status"               db:"status"` // "trialing"|"active"|"past_due"|"canceled"|"expired"
	CurrentPeriodStart time.Time  `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"   db:"current_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	CreatedAt          time.Time  `json:"created_at"           db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"           db:"updated_at"`
}

type SubscriptionWithTariff struct {
	Subscription
	TariffName     string         `json:"tariff_name"     db:"tariff_name"`
	TariffSlug     string         `json:"tariff_slug"     db:"tariff_slug"`
	TariffPrice    int            `json:"tariff_price"    db:"tariff_price"`
	TariffFeatures TariffFeatures `json:"tariff_features" db:"tariff_features"`
	TariffLimits   TariffLimits   `json:"tariff_limits"   db:"tariff_limits"`
}

type Invoice struct {
	ID             int        `json:"id"              db:"id"`
	OrgID          int        `json:"org_id"          db:"org_id"`
	SubscriptionID *int       `json:"subscription_id,omitempty" db:"subscription_id"`
	Amount         int        `json:"amount"          db:"amount"` // kopeks
	Currency       string     `json:"currency"        db:"currency"`
	Status         string     `json:"status"          db:"status"` // "pending"|"paid"|"failed"|"refunded"
	DueDate        time.Time  `json:"due_date"        db:"due_date"`
	PaidAt         *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt      time.Time  `json:"created_at"      db:"created_at"`
}

type Payment struct {
	ID                int       `json:"id"                  db:"id"`
	InvoiceID         int       `json:"invoice_id"          db:"invoice_id"`
	OrgID             int       `json:"org_id"              db:"org_id"`
	Amount            int       `json:"amount"              db:"amount"` // kopeks
	Currency          string    `json:"currency"            db:"currency"`
	Provider          string    `json:"provider"            db:"provider"`
	ProviderPaymentID *string   `json:"provider_payment_id,omitempty" db:"provider_payment_id"`
	Status            string    `json:"status"              db:"status"` // "pending"|"succeeded"|"failed"|"refunded"
	CreatedAt         time.Time `json:"created_at"          db:"created_at"`
}

// Request DTOs

type CreateSubscriptionRequest struct {
	TariffSlug string `json:"tariff_slug" binding:"required"`
}

type ChangeSubscriptionRequest struct {
	TariffSlug string `json:"tariff_slug" binding:"required"`
}

type ProcessPaymentRequest struct {
	InvoiceID         int    `json:"invoice_id" binding:"required"`
	Provider          string `json:"provider" binding:"required"`
	ProviderPaymentID string `json:"provider_payment_id"`
	Amount            int    `json:"amount" binding:"required"`
}
