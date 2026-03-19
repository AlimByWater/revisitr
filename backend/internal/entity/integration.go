package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type IntegrationConfig struct {
	APIURL string `json:"api_url,omitempty"`
	APIKey string `json:"api_key,omitempty"`
}

func (c IntegrationConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("IntegrationConfig.Value: %w", err)
	}
	return b, nil
}

func (c *IntegrationConfig) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = IntegrationConfig{}
		return nil
	default:
		return fmt.Errorf("IntegrationConfig.Scan: unsupported type %T", src)
	}
}

type Integration struct {
	ID         int               `json:"id"          db:"id"`
	OrgID      int               `json:"org_id"      db:"org_id"`
	Type       string            `json:"type"        db:"type"` // "iiko"|"rkeeper"|"1c"
	Config     IntegrationConfig `json:"config"      db:"config"`
	Status     string            `json:"status"      db:"status"` // "active"|"inactive"|"error"
	LastSyncAt *time.Time        `json:"last_sync_at,omitempty" db:"last_sync_at"`
	CreatedAt  time.Time         `json:"created_at"  db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"  db:"updated_at"`
}

type CreateIntegrationRequest struct {
	Type   string            `json:"type"   binding:"required,oneof=iiko rkeeper 1c"`
	Config IntegrationConfig `json:"config" binding:"required"`
}

type UpdateIntegrationRequest struct {
	Config *IntegrationConfig `json:"config,omitempty"`
	Status *string            `json:"status,omitempty"`
}

// OrderItem — a single line item in an external order.
type OrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type OrderItems []OrderItem

func (items OrderItems) Value() (driver.Value, error) {
	b, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("OrderItems.Value: %w", err)
	}
	return b, nil
}

func (items *OrderItems) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, items)
	case string:
		return json.Unmarshal([]byte(v), items)
	case nil:
		*items = OrderItems{}
		return nil
	default:
		return fmt.Errorf("OrderItems.Scan: unsupported type %T", src)
	}
}

type ExternalOrder struct {
	ID            int        `json:"id"             db:"id"`
	IntegrationID int        `json:"integration_id" db:"integration_id"`
	ExternalID    string     `json:"external_id"    db:"external_id"`
	ClientID      *int       `json:"client_id,omitempty" db:"client_id"`
	Items         OrderItems `json:"items"          db:"items"`
	Total         float64    `json:"total"          db:"total"`
	OrderedAt     *time.Time `json:"ordered_at,omitempty" db:"ordered_at"`
	SyncedAt      time.Time  `json:"synced_at"      db:"synced_at"`
}
