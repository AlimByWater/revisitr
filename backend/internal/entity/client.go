package entity

type ClientFilter struct {
	BotID     *int    `form:"bot_id"`
	Segment   *string `form:"segment"`
	Search    *string `form:"search"`
	SortBy    string  `form:"sort_by"`
	SortOrder string  `form:"sort_order"`
	Limit     int     `form:"limit"`
	Offset    int     `form:"offset"`
}

type ClientProfile struct {
	BotClient
	BotName        string               `db:"bot_name" json:"bot_name"`
	LoyaltyBalance float64              `db:"loyalty_balance" json:"loyalty_balance"`
	LoyaltyLevel   *string              `db:"loyalty_level" json:"loyalty_level"`
	TotalPurchases float64              `db:"total_purchases" json:"total_purchases"`
	PurchaseCount  int                  `db:"purchase_count" json:"purchase_count"`
	Transactions   []LoyaltyTransaction `db:"-" json:"transactions,omitempty"`
}

type ClientStats struct {
	TotalClients   int     `db:"total_clients" json:"total_clients"`
	TotalBalance   float64 `db:"total_balance" json:"total_balance"`
	NewThisMonth   int     `db:"new_this_month" json:"new_this_month"`
	ActiveThisWeek int     `db:"active_this_week" json:"active_this_week"`
}

type UpdateClientRequest struct {
	Tags *Tags `json:"tags,omitempty"`
}

type PaginatedResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}
