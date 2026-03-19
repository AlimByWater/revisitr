package promotions

import "errors"

var (
	ErrPromotionNotFound    = errors.New("promotion not found")
	ErrNotPromotionOwner    = errors.New("not authorized")
	ErrPromoCodeNotFound    = errors.New("promo code not found")
	ErrPromoCodeInactive    = errors.New("promo code is inactive")
	ErrPromoCodeNotActive   = errors.New("promo code is not yet active")
	ErrPromoCodeExpired     = errors.New("promo code has expired")
	ErrPromoCodeLimitReached = errors.New("promo code usage limit reached")
)
