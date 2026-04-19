package entity

import "time"

type EmojiPack struct {
	ID        int         `db:"id"         json:"id"`
	OrgID     int         `db:"org_id"     json:"org_id"`
	Name      string      `db:"name"       json:"name"`
	SortOrder int         `db:"sort_order" json:"sort_order"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
	Items     []EmojiItem `db:"-"          json:"items,omitempty"`
}

type EmojiItem struct {
	ID              int       `db:"id"                 json:"id"`
	PackID          int       `db:"pack_id"            json:"pack_id"`
	Name            string    `db:"name"               json:"name"`
	ImageURL        string    `db:"image_url"          json:"image_url"`
	SortOrder       int       `db:"sort_order"         json:"sort_order"`
	TgStickerSet    *string   `db:"tg_sticker_set"     json:"tg_sticker_set,omitempty"`
	TgCustomEmojiID *string   `db:"tg_custom_emoji_id" json:"tg_custom_emoji_id,omitempty"`
	CreatedAt       time.Time `db:"created_at"         json:"created_at"`
}

type CreateEmojiPackRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateEmojiPackRequest struct {
	Name      *string `json:"name,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

type CreateEmojiItemRequest struct {
	Name     string `json:"name"      binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
}

type UpdateEmojiItemRequest struct {
	Name      *string `json:"name,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

type ReorderEmojiItemsRequest struct {
	ItemIDs []int `json:"item_ids" binding:"required"`
}
