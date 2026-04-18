package entity

import "time"

type Menu struct {
	ID            int              `db:"id"             json:"id"`
	OrgID         int              `db:"org_id"         json:"org_id"`
	IntegrationID *int             `db:"integration_id" json:"integration_id,omitempty"`
	Name          string           `db:"name"           json:"name"`
	Source        string           `db:"source"         json:"source"` // "manual"|"pos_import"
	IntroContent  *MessageContent  `db:"intro_content" json:"intro_content,omitempty"`
	LastSyncedAt  *time.Time       `db:"last_synced_at" json:"last_synced_at,omitempty"`
	CreatedAt     time.Time        `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at"     json:"updated_at"`
	Categories    []MenuCategory   `db:"-" json:"categories,omitempty"`
	Bindings      []MenuPOSBinding `db:"-" json:"bindings,omitempty"`
}

type MenuCategory struct {
	ID           int        `db:"id"             json:"id"`
	MenuID       int        `db:"menu_id"        json:"menu_id"`
	Name         string     `db:"name"           json:"name"`
	IconEmoji    *string    `db:"icon_emoji"     json:"icon_emoji,omitempty"`
	IconImageURL *string    `db:"icon_image_url" json:"icon_image_url,omitempty"`
	SortOrder    int        `db:"sort_order"     json:"sort_order"`
	CreatedAt    time.Time  `db:"created_at"     json:"created_at"`
	Items        []MenuItem `db:"-"              json:"items,omitempty"`
}

type MenuItem struct {
	ID          int       `db:"id"           json:"id"`
	CategoryID  int       `db:"category_id"  json:"category_id"`
	Name        string    `db:"name"         json:"name"`
	Description *string   `db:"description"  json:"description,omitempty"`
	Price       float64   `db:"price"        json:"price"`
	Weight      *string   `db:"weight"       json:"weight,omitempty"`
	ImageURL    *string   `db:"image_url"    json:"image_url,omitempty"`
	Tags        Tags      `db:"tags"         json:"tags"`
	ExternalID  *string   `db:"external_id"  json:"external_id,omitempty"`
	IsAvailable bool      `db:"is_available" json:"is_available"`
	SortOrder   int       `db:"sort_order"   json:"sort_order"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

type MenuPOSBinding struct {
	MenuID    int       `db:"menu_id" json:"menu_id"`
	POSID     int       `db:"pos_id" json:"pos_id"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	POSName   string    `db:"pos_name" json:"pos_name,omitempty"`
}

type CreateMenuRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateMenuRequest struct {
	Name         *string                 `json:"name,omitempty"`
	IntroContent *MessageContent         `json:"intro_content,omitempty"`
	Bindings     []MenuPOSBindingRequest `json:"bindings,omitempty"`
}

type CreateMenuCategoryRequest struct {
	Name         string `json:"name" binding:"required"`
	IconEmoji    string `json:"icon_emoji,omitempty"`
	IconImageURL string `json:"icon_image_url,omitempty"`
	SortOrder    int    `json:"sort_order"`
}

type UpdateMenuCategoryRequest struct {
	Name         *string `json:"name,omitempty"`
	IconEmoji    *string `json:"icon_emoji,omitempty"`
	IconImageURL *string `json:"icon_image_url,omitempty"`
	SortOrder    *int    `json:"sort_order,omitempty"`
}

type CreateMenuItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"min=0"`
	Weight      string  `json:"weight,omitempty"`
	ImageURL    string  `json:"image_url"`
	Tags        Tags    `json:"tags"`
}

type UpdateMenuItemRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Weight      *string  `json:"weight,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
	Tags        *Tags    `json:"tags,omitempty"`
	IsAvailable *bool    `json:"is_available,omitempty"`
	SortOrder   *int     `json:"sort_order,omitempty"`
}

type MenuPOSBindingRequest struct {
	POSID    int  `json:"pos_id"`
	IsActive bool `json:"is_active"`
}

// ClientOrderStats — aggregated order data for a client.
type ClientOrderStats struct {
	TotalOrders int            `json:"total_orders" db:"total_orders"`
	TotalAmount float64        `json:"total_amount" db:"total_amount"`
	AvgAmount   float64        `json:"avg_amount"   db:"avg_amount"`
	LastOrderAt *time.Time     `json:"last_order_at,omitempty" db:"last_order_at"`
	TopItems    []TopOrderItem `json:"top_items" db:"-"`
}

type TopOrderItem struct {
	Name       string  `json:"name"        db:"name"`
	OrderCount int     `json:"order_count" db:"order_count"`
	TotalQty   int     `json:"total_qty"   db:"total_qty"`
	TotalSum   float64 `json:"total_sum"   db:"total_sum"`
}
