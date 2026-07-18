package entity

import "time"

const (
	OrderSourceLunch = "lunch"
	OrderSourceMenu  = "menu"

	OrderStatusNew       = "new"
	OrderStatusSent      = "sent"
	OrderStatusCancelled = "cancelled"
)

// Order — заказ гостя, оформленный через бота. Source указывает, какой
// флоу создал заказ (lunch, menu). Не путать с external_orders (импорт из POS).
type Order struct {
	ID          int       `db:"id"            json:"id"`
	BotID       int       `db:"bot_id"        json:"bot_id"`
	BotClientID int       `db:"bot_client_id" json:"bot_client_id"`
	Source      string    `db:"source"        json:"source"`
	FormatID    *int      `db:"format_id"     json:"format_id,omitempty"` // lunch only
	FormatName  string    `db:"format_name"   json:"format_name"`         // lunch only
	TableNum    string    `db:"table_num"     json:"table_num"`
	TotalPrice  float64   `db:"total_price"   json:"total_price"`
	Status      string    `db:"status"        json:"status"`
	CreatedAt   time.Time `db:"created_at"    json:"created_at"`

	// BotName заполняется только в org-wide выборках (JOIN bots).
	BotName string      `db:"bot_name" json:"bot_name,omitempty"`
	Items   []OrderLine `db:"-"        json:"items"`
}

// OrderLine — позиция заказа со снапшотами названия и цены.
// Имя OrderItem занято POS-концептом (external_orders, integration.go).
type OrderLine struct {
	ID          int     `db:"id"           json:"id"`
	OrderID     int     `db:"order_id"     json:"order_id"`
	CourseID    *int    `db:"course_id"    json:"course_id,omitempty"` // lunch only
	CourseTitle string  `db:"course_title" json:"course_title"`        // lunch only
	MenuItemID  *int    `db:"menu_item_id" json:"menu_item_id,omitempty"`
	ItemName    string  `db:"item_name"    json:"item_name"`
	Price       float64 `db:"price"        json:"price"`
	Surcharge   float64 `db:"surcharge"    json:"surcharge"` // lunch only
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
