package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

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
}

type CreatePromotionRequest struct {
	Name       string              `json:"name"        binding:"required"`
	Type       string              `json:"type"        binding:"required,oneof=discount bonus tag_update campaign"`
	Conditions PromotionConditions `json:"conditions"`
	Result     PromotionResult     `json:"result"      binding:"required"`
	StartsAt   *time.Time          `json:"starts_at,omitempty"`
	EndsAt     *time.Time          `json:"ends_at,omitempty"`
	UsageLimit *int                `json:"usage_limit,omitempty"`
	Recurrence string              `json:"recurrence" binding:"omitempty,oneof=one_time daily weekly monthly"`
	Combinable bool                `json:"combinable"`
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
