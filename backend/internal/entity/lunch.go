package entity

import "time"

const (
	LunchPriceFixed             = "fixed"
	LunchPriceSumOfItems        = "sum_of_items"
	LunchPriceBasePlusSurcharge = "base_plus_surcharge"
)

type LunchProgram struct {
	ID          int       `db:"id"          json:"id"`
	BotID       int       `db:"bot_id"      json:"bot_id"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description"`
	IsActive    bool      `db:"is_active"   json:"is_active"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`

	Courses      []LunchCourse       `db:"-" json:"courses"`
	Formats      []LunchFormat       `db:"-" json:"formats"`
	Availability []LunchAvailability `db:"-" json:"availability"`
}

type LunchCourse struct {
	ID             int    `db:"id"               json:"id"`
	ProgramID      int    `db:"program_id"       json:"program_id"`
	Code           string `db:"code"             json:"code"`
	Title          string `db:"title"            json:"title"`
	MenuCategoryID int    `db:"menu_category_id" json:"menu_category_id"`
	SortOrder      int    `db:"sort_order"       json:"sort_order"`

	Items []LunchCourseItem `db:"-" json:"items"`
}

type LunchCourseItem struct {
	CourseID   int     `db:"course_id"    json:"course_id"`
	MenuItemID int     `db:"menu_item_id" json:"menu_item_id"`
	Surcharge  float64 `db:"surcharge"    json:"surcharge"`

	MenuItem *MenuItem `db:"-" json:"menu_item,omitempty"`
}

type LunchFormat struct {
	ID        int     `db:"id"         json:"id"`
	ProgramID int     `db:"program_id" json:"program_id"`
	Name      string  `db:"name"       json:"name"`
	PriceMode string  `db:"price_mode" json:"price_mode"`
	BasePrice float64 `db:"base_price" json:"base_price"`
	IsActive  bool    `db:"is_active"  json:"is_active"`
	SortOrder int     `db:"sort_order" json:"sort_order"`

	CourseIDs []int `db:"-" json:"course_ids"` // ordered by lunch_format_courses.position
}

type LunchAvailability struct {
	ID        int    `db:"id"         json:"id"`
	ProgramID int    `db:"program_id" json:"program_id"`
	Weekday   int    `db:"weekday"    json:"weekday"` // ISO: 1 = Monday … 7 = Sunday
	TimeFrom  string `db:"time_from"  json:"time_from"` // "12:00"
	TimeTo    string `db:"time_to"    json:"time_to"`
}

type UpsertLunchProgramRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type LunchCourseItemRequest struct {
	MenuItemID int     `json:"menu_item_id" binding:"required"`
	Surcharge  float64 `json:"surcharge"`
}

type CreateLunchCourseRequest struct {
	Code           string                   `json:"code" binding:"required"`
	Title          string                   `json:"title" binding:"required"`
	MenuCategoryID int                      `json:"menu_category_id" binding:"required"`
	SortOrder      int                      `json:"sort_order"`
	Items          []LunchCourseItemRequest `json:"items"`
}

type UpdateLunchCourseRequest struct {
	Code           *string                  `json:"code,omitempty"`
	Title          *string                  `json:"title,omitempty"`
	MenuCategoryID *int                     `json:"menu_category_id,omitempty"`
	SortOrder      *int                     `json:"sort_order,omitempty"`
	Items          []LunchCourseItemRequest `json:"items,omitempty"`
}

type CreateLunchFormatRequest struct {
	Name      string  `json:"name" binding:"required"`
	PriceMode string  `json:"price_mode" binding:"required"`
	BasePrice float64 `json:"base_price"`
	IsActive  bool    `json:"is_active"`
	SortOrder int     `json:"sort_order"`
	CourseIDs []int   `json:"course_ids" binding:"required"`
}

type UpdateLunchFormatRequest struct {
	Name      *string  `json:"name,omitempty"`
	PriceMode *string  `json:"price_mode,omitempty"`
	BasePrice *float64 `json:"base_price,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
	SortOrder *int     `json:"sort_order,omitempty"`
	CourseIDs []int    `json:"course_ids,omitempty"`
}

type SetLunchAvailabilityRequest struct {
	Slots []LunchAvailability `json:"slots"`
}
