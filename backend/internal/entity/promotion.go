package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ── Promotion Triggers & Actions (v2 wizard model) ──────────────────────────

// PromotionTrigger represents a trigger condition for a promotion.
type PromotionTrigger struct {
	Type      string   `json:"type"`                    // purchase, purchase_product, purchase_min_items, receipt_sum, event
	ProductID *int     `json:"product_id,omitempty"`    // for purchase_product
	MinItems  *int     `json:"min_items,omitempty"`     // for purchase_min_items
	MinAmount *float64 `json:"min_amount,omitempty"`    // for receipt_sum
	EventType *string  `json:"event_type,omitempty"`    // birthday, registration, activation, last_purchase
}

// PromotionTriggers is a slice of PromotionTrigger stored as JSONB.
type PromotionTriggers []PromotionTrigger

func (t PromotionTriggers) Value() (driver.Value, error) {
	if t == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("PromotionTriggers.Value: %w", err)
	}
	return b, nil
}

func (t *PromotionTriggers) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	case nil:
		*t = nil
		return nil
	default:
		return fmt.Errorf("PromotionTriggers.Scan: unsupported type %T", src)
	}
}

// PromotionAction represents an action to execute when a promotion triggers.
type PromotionAction struct {
	Type            string   `json:"type"`                       // discount, bonus, data_update, campaign
	DiscountPercent *float64 `json:"discount_percent,omitempty"` // for discount
	BonusAmount     *int     `json:"bonus_amount,omitempty"`     // for bonus
	TagAdd          *string  `json:"tag_add,omitempty"`          // for data_update
	LevelID         *int     `json:"level_id,omitempty"`         // for data_update
	CampaignID      *int     `json:"campaign_id,omitempty"`      // for campaign
}

// PromotionActions is a slice of PromotionAction stored as JSONB.
type PromotionActions []PromotionAction

func (a PromotionActions) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("PromotionActions.Value: %w", err)
	}
	return b, nil
}

func (a *PromotionActions) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	case nil:
		*a = nil
		return nil
	default:
		return fmt.Errorf("PromotionActions.Scan: unsupported type %T", src)
	}
}

// ── Legacy conditions/result (v1 model, kept for backward compat) ───────────

type PromotionConditions struct {
	MinAmount  *float64 `json:"min_amount,omitempty"`
	SegmentID  *int     `json:"segment_id,omitempty"`
	MinVisits  *int     `json:"min_visits,omitempty"`
}

func (c PromotionConditions) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("PromotionConditions.Value: %w", err)
	}
	return b, nil
}

func (c *PromotionConditions) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = PromotionConditions{}
		return nil
	default:
		return fmt.Errorf("PromotionConditions.Scan: unsupported type %T", src)
	}
}

type PromotionResult struct {
	DiscountPercent *float64 `json:"discount_percent,omitempty"`
	BonusAmount     *int     `json:"bonus_amount,omitempty"`
	TagAdd          *string  `json:"tag_add,omitempty"`
	CampaignID      *int     `json:"campaign_id,omitempty"`
}

func (r PromotionResult) Value() (driver.Value, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("PromotionResult.Value: %w", err)
	}
	return b, nil
}

func (r *PromotionResult) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, r)
	case string:
		return json.Unmarshal([]byte(v), r)
	case nil:
		*r = PromotionResult{}
		return nil
	default:
		return fmt.Errorf("PromotionResult.Scan: unsupported type %T", src)
	}
}

type Promotion struct {
	ID         int                 `json:"id"          db:"id"`
	OrgID      int                 `json:"org_id"      db:"org_id"`
	Name       string              `json:"name"        db:"name"`
	Type       string              `json:"type"        db:"type"` // "discount"|"bonus"|"tag_update"|"campaign"
	Conditions PromotionConditions `json:"conditions"  db:"conditions"`
	Result     PromotionResult     `json:"result"      db:"result"`
	StartsAt   *time.Time          `json:"starts_at,omitempty"   db:"starts_at"`
	EndsAt     *time.Time          `json:"ends_at,omitempty"     db:"ends_at"`
	UsageLimit *int                `json:"usage_limit,omitempty" db:"usage_limit"`
	Recurrence string              `json:"recurrence"  db:"recurrence"` // "one_time"|"daily"|"weekly"|"monthly"
	Combinable bool                `json:"combinable"  db:"combinable"`
	Active     bool                `json:"active"      db:"active"`
	CreatedAt  time.Time           `json:"created_at"  db:"created_at"`

	// v2 wizard fields
	Filter               SegmentFilter     `json:"filter"                  db:"filter"`
	Triggers             PromotionTriggers `json:"triggers"                db:"triggers"`
	Actions              PromotionActions  `json:"actions"                 db:"actions"`
	CombinableWithLoyalty bool             `json:"combinable_with_loyalty" db:"combinable_with_loyalty"`
}

type CreatePromotionRequest struct {
	Name       string              `json:"name"        binding:"required"`
	Type       string              `json:"type"        binding:"required,oneof=discount bonus tag_update campaign"`
	Conditions PromotionConditions `json:"conditions"`
	Result     PromotionResult     `json:"result"`
	StartsAt   *time.Time          `json:"starts_at,omitempty"`
	EndsAt     *time.Time          `json:"ends_at,omitempty"`
	UsageLimit *int                `json:"usage_limit,omitempty"`
	Recurrence string              `json:"recurrence" binding:"omitempty,oneof=one_time daily weekly monthly"`
	Combinable bool                `json:"combinable"`

	// v2 wizard fields
	Filter                SegmentFilter     `json:"filter"`
	Triggers              PromotionTriggers `json:"triggers"`
	Actions               PromotionActions  `json:"actions"`
	CombinableWithLoyalty bool              `json:"combinable_with_loyalty"`
}

type UpdatePromotionRequest struct {
	Name       *string              `json:"name,omitempty"`
	Conditions *PromotionConditions `json:"conditions,omitempty"`
	Result     *PromotionResult     `json:"result,omitempty"`
	StartsAt   *time.Time           `json:"starts_at,omitempty"`
	EndsAt     *time.Time           `json:"ends_at,omitempty"`
	UsageLimit *int                 `json:"usage_limit,omitempty"`
	Combinable *bool                `json:"combinable,omitempty"`
	Active     *bool                `json:"active,omitempty"`

	// v2 wizard fields
	Filter                *SegmentFilter     `json:"filter,omitempty"`
	Triggers              *PromotionTriggers `json:"triggers,omitempty"`
	Actions               *PromotionActions  `json:"actions,omitempty"`
	CombinableWithLoyalty *bool              `json:"combinable_with_loyalty,omitempty"`
}

// PromoCodeConditions — conditions for a promo code.
type PromoCodeConditions struct {
	MinAmount *float64 `json:"min_amount,omitempty"`
}

func (c PromoCodeConditions) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("PromoCodeConditions.Value: %w", err)
	}
	return b, nil
}

func (c *PromoCodeConditions) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = PromoCodeConditions{}
		return nil
	default:
		return fmt.Errorf("PromoCodeConditions.Scan: unsupported type %T", src)
	}
}

type PromoCode struct {
	ID              int                 `json:"id"               db:"id"`
	OrgID           int                 `json:"org_id"           db:"org_id"`
	PromotionID     *int                `json:"promotion_id"     db:"promotion_id"`
	Code            string              `json:"code"             db:"code"`
	DiscountPercent *float64            `json:"discount_percent" db:"discount_percent"`
	BonusAmount     *int                `json:"bonus_amount"     db:"bonus_amount"`
	Channel         *string             `json:"channel"          db:"channel"`
	PerUserLimit    *int                `json:"per_user_limit"   db:"per_user_limit"`
	Description     *string             `json:"description"      db:"description"`
	StartsAt        *time.Time          `json:"starts_at,omitempty" db:"starts_at"`
	EndsAt          *time.Time          `json:"ends_at,omitempty"   db:"ends_at"`
	Conditions      PromoCodeConditions `json:"conditions"       db:"conditions"`
	UsageCount      int                 `json:"usage_count"      db:"usage_count"`
	UsageLimit      *int                `json:"usage_limit,omitempty" db:"usage_limit"`
	Active          bool                `json:"active"           db:"active"`
	CreatedAt       time.Time           `json:"created_at"       db:"created_at"`
}

type CreatePromoCodeRequest struct {
	PromotionID     *int                `json:"promotion_id,omitempty"`
	Code            string              `json:"code"             binding:"required"`
	DiscountPercent *float64            `json:"discount_percent,omitempty"`
	BonusAmount     *int                `json:"bonus_amount,omitempty"`
	Channel         *string             `json:"channel,omitempty"`
	PerUserLimit    *int                `json:"per_user_limit,omitempty"`
	Description     *string             `json:"description,omitempty"`
	StartsAt        *time.Time          `json:"starts_at,omitempty"`
	EndsAt          *time.Time          `json:"ends_at,omitempty"`
	Conditions      PromoCodeConditions `json:"conditions"`
	UsageLimit      *int                `json:"usage_limit,omitempty"`
}

// PromoResult — result of applying a promo code.
type PromoResult struct {
	Code            string   `json:"code"`
	DiscountPercent *float64 `json:"discount_percent,omitempty"`
	BonusAmount     *int     `json:"bonus_amount,omitempty"`
}

// PromoChannelAnalytics — analytics from promo_channel_analytics view.
type PromoChannelAnalytics struct {
	Channel       string `json:"channel"        db:"channel"`
	CodeCount     int    `json:"code_count"     db:"code_count"`
	TotalUsages   int    `json:"total_usages"   db:"total_usages"`
	UniqueClients int    `json:"unique_clients" db:"unique_clients"`
}

// PromoCodeValidation — result of promo code validation.
type PromoCodeValidation struct {
	Valid  bool         `json:"valid"`
	Reason string      `json:"reason,omitempty"`
	Promo  *PromoResult `json:"promo,omitempty"`
}
