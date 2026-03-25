package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type MarketplaceProduct struct {
	ID          int       `db:"id" json:"id"`
	OrgID       int       `db:"org_id" json:"org_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	ImageURL    string    `db:"image_url" json:"image_url"`
	PricePoints int       `db:"price_points" json:"price_points"`
	Stock       *int      `db:"stock" json:"stock"` // nil = unlimited
	IsActive    bool      `db:"is_active" json:"is_active"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type MarketplaceOrderItem struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Points      int    `json:"points"`
}

type MarketplaceOrderItems []MarketplaceOrderItem

func (o MarketplaceOrderItems) Value() (driver.Value, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("MarketplaceOrderItems.Value: %w", err)
	}
	return b, nil
}

func (o *MarketplaceOrderItems) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, o)
	case string:
		return json.Unmarshal([]byte(v), o)
	case nil:
		*o = MarketplaceOrderItems{}
		return nil
	default:
		return fmt.Errorf("MarketplaceOrderItems.Scan: unsupported type %T", src)
	}
}

type MarketplaceOrder struct {
	ID          int        `db:"id" json:"id"`
	OrgID       int        `db:"org_id" json:"org_id"`
	ClientID    int        `db:"client_id" json:"client_id"`
	Status      string     `db:"status" json:"status"`
	TotalPoints int        `db:"total_points" json:"total_points"`
	Items       MarketplaceOrderItems `db:"items" json:"items"`
	Note        string     `db:"note" json:"note"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// --- Request DTOs ---

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	PricePoints int    `json:"price_points" binding:"required,gt=0"`
	Stock       *int   `json:"stock"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
	PricePoints *int    `json:"price_points"`
	Stock       *int    `json:"stock"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

type PlaceOrderRequest struct {
	ClientID  int              `json:"client_id" binding:"required"`
	ProgramID int              `json:"program_id" binding:"required"`
	Items     []PlaceOrderItem `json:"items" binding:"required,min=1"`
	Note      string           `json:"note"`
}

type PlaceOrderItem struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,gt=0"`
}

type MarketplaceStats struct {
	TotalProducts  int `json:"total_products"`
	ActiveProducts int `json:"active_products"`
	TotalOrders    int `json:"total_orders"`
	TotalSpent     int `json:"total_spent_points"`
}
