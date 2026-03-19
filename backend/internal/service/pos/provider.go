package pos

import (
	"context"
	"time"
)

// POSProvider is the common interface for all POS system integrations.
type POSProvider interface {
	TestConnection(ctx context.Context) error
	GetCustomers(ctx context.Context, opts CustomerListOpts) ([]POSCustomer, error)
	GetOrders(ctx context.Context, from, to time.Time) ([]POSOrder, error)
	GetMenu(ctx context.Context) (*POSMenu, error)
}

type POSCustomer struct {
	ExternalID string     `json:"external_id"`
	Phone      string     `json:"phone"`
	Name       string     `json:"name"`
	Email      string     `json:"email,omitempty"`
	Birthday   *time.Time `json:"birthday,omitempty"`
	Balance    float64    `json:"balance"`
	CardNumber string     `json:"card_number,omitempty"`
}

type POSOrder struct {
	ExternalID    string         `json:"external_id"`
	CustomerPhone string         `json:"customer_phone,omitempty"`
	Items         []POSOrderItem `json:"items"`
	Total         float64        `json:"total"`
	Discount      float64        `json:"discount"`
	OrderedAt     time.Time      `json:"ordered_at"`
	Status        string         `json:"status"` // "open", "closed", "cancelled"
	TableNum      string         `json:"table_num,omitempty"`
	WaiterName    string         `json:"waiter_name,omitempty"`
}

type POSOrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Category string  `json:"category,omitempty"`
}

type POSMenu struct {
	Categories []MenuCategory `json:"categories"`
}

type MenuCategory struct {
	Name  string     `json:"name"`
	Items []MenuItem `json:"items"`
}

type MenuItem struct {
	ExternalID  string  `json:"external_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
}

type CustomerListOpts struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Search string `json:"search,omitempty"`
}
