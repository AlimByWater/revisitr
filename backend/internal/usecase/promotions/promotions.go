package promotions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type promotionsRepo interface {
	Create(ctx context.Context, p *entity.Promotion) error
	GetByID(ctx context.Context, id int) (*entity.Promotion, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Promotion, error)
	Update(ctx context.Context, p *entity.Promotion) error
	Delete(ctx context.Context, id int) error
	GetApplicable(ctx context.Context, orgID, clientID int) ([]entity.Promotion, error)
	RecordUsage(ctx context.Context, promotionID, clientID int, promoCodeID *int) error
	CreatePromoCode(ctx context.Context, pc *entity.PromoCode) error
	GetPromoCodeByCode(ctx context.Context, orgID int, code string) (*entity.PromoCode, error)
	GetPromoCodeByID(ctx context.Context, id int) (*entity.PromoCode, error)
	ListPromoCodes(ctx context.Context, orgID int) ([]entity.PromoCode, error)
	IncrementPromoCodeUsage(ctx context.Context, id int) error
	DeactivatePromoCode(ctx context.Context, id int) error
}

type Usecase struct {
	logger     *slog.Logger
	promotions promotionsRepo
}

func New(promotions promotionsRepo) *Usecase {
	return &Usecase{promotions: promotions}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// --- Promotions CRUD ---

func (uc *Usecase) Create(ctx context.Context, orgID int, req *entity.CreatePromotionRequest) (*entity.Promotion, error) {
	p := &entity.Promotion{
		OrgID:      orgID,
		Name:       req.Name,
		Type:       req.Type,
		Conditions: req.Conditions,
		Result:     req.Result,
		StartsAt:   req.StartsAt,
		EndsAt:     req.EndsAt,
		UsageLimit: req.UsageLimit,
		Combinable: req.Combinable,
		Active:     true,
	}

	if err := uc.promotions.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create promotion: %w", err)
	}

	return p, nil
}

func (uc *Usecase) List(ctx context.Context, orgID int) ([]entity.Promotion, error) {
	promos, err := uc.promotions.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list promotions: %w", err)
	}
	return promos, nil
}

func (uc *Usecase) GetByID(ctx context.Context, id, orgID int) (*entity.Promotion, error) {
	p, err := uc.promotions.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPromotionNotFound
		}
		return nil, fmt.Errorf("get promotion: %w", err)
	}

	if p.OrgID != orgID {
		return nil, ErrNotPromotionOwner
	}

	return p, nil
}

func (uc *Usecase) Update(ctx context.Context, id, orgID int, req *entity.UpdatePromotionRequest) (*entity.Promotion, error) {
	p, err := uc.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Conditions != nil {
		p.Conditions = *req.Conditions
	}
	if req.Result != nil {
		p.Result = *req.Result
	}
	if req.StartsAt != nil {
		p.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		p.EndsAt = req.EndsAt
	}
	if req.UsageLimit != nil {
		p.UsageLimit = req.UsageLimit
	}
	if req.Combinable != nil {
		p.Combinable = *req.Combinable
	}
	if req.Active != nil {
		p.Active = *req.Active
	}

	if err := uc.promotions.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update promotion: %w", err)
	}

	return p, nil
}

func (uc *Usecase) Delete(ctx context.Context, id, orgID int) error {
	if _, err := uc.GetByID(ctx, id, orgID); err != nil {
		return err
	}

	if err := uc.promotions.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete promotion: %w", err)
	}

	return nil
}

// --- Promo codes ---

func (uc *Usecase) CreatePromoCode(ctx context.Context, orgID int, req *entity.CreatePromoCodeRequest) (*entity.PromoCode, error) {
	pc := &entity.PromoCode{
		OrgID:           orgID,
		PromotionID:     req.PromotionID,
		Code:            req.Code,
		DiscountPercent: req.DiscountPercent,
		BonusAmount:     req.BonusAmount,
		StartsAt:        req.StartsAt,
		EndsAt:          req.EndsAt,
		Conditions:      req.Conditions,
		UsageLimit:      req.UsageLimit,
		Active:          true,
	}

	if err := uc.promotions.CreatePromoCode(ctx, pc); err != nil {
		return nil, fmt.Errorf("create promo code: %w", err)
	}

	return pc, nil
}

func (uc *Usecase) ListPromoCodes(ctx context.Context, orgID int) ([]entity.PromoCode, error) {
	codes, err := uc.promotions.ListPromoCodes(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list promo codes: %w", err)
	}
	return codes, nil
}

func (uc *Usecase) DeactivatePromoCode(ctx context.Context, id, orgID int) error {
	pc, err := uc.promotions.GetPromoCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPromoCodeNotFound
		}
		return fmt.Errorf("get promo code: %w", err)
	}

	if pc.OrgID != orgID {
		return ErrNotPromotionOwner
	}

	if err := uc.promotions.DeactivatePromoCode(ctx, id); err != nil {
		return fmt.Errorf("deactivate promo code: %w", err)
	}

	return nil
}

// ApplyPromoCode validates and applies a promo code for a client.
func (uc *Usecase) ApplyPromoCode(ctx context.Context, orgID, clientID int, code string) (*entity.PromoResult, error) {
	pc, err := uc.promotions.GetPromoCodeByCode(ctx, orgID, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPromoCodeNotFound
		}
		return nil, fmt.Errorf("get promo code: %w", err)
	}

	if !pc.Active {
		return nil, ErrPromoCodeInactive
	}

	now := time.Now()
	if pc.StartsAt != nil && now.Before(*pc.StartsAt) {
		return nil, ErrPromoCodeNotActive
	}
	if pc.EndsAt != nil && now.After(*pc.EndsAt) {
		return nil, ErrPromoCodeExpired
	}
	if pc.UsageLimit != nil && pc.UsageCount >= *pc.UsageLimit {
		return nil, ErrPromoCodeLimitReached
	}

	if err := uc.promotions.IncrementPromoCodeUsage(ctx, pc.ID); err != nil {
		return nil, fmt.Errorf("increment promo code usage: %w", err)
	}

	if err := uc.promotions.RecordUsage(ctx, 0, clientID, &pc.ID); err != nil {
		uc.logger.Error("record promo code usage", "error", err, "promo_code_id", pc.ID)
	}

	return &entity.PromoResult{
		Code:            pc.Code,
		DiscountPercent: pc.DiscountPercent,
		BonusAmount:     pc.BonusAmount,
	}, nil
}

// GetApplicable returns active promotions applicable for a given client.
func (uc *Usecase) GetApplicable(ctx context.Context, orgID, clientID int) ([]entity.Promotion, error) {
	promos, err := uc.promotions.GetApplicable(ctx, orgID, clientID)
	if err != nil {
		return nil, fmt.Errorf("get applicable promotions: %w", err)
	}
	return promos, nil
}
