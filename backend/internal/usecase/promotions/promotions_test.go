package promotions

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockPromotionsRepo struct {
	createFn                  func(ctx context.Context, p *entity.Promotion) error
	getByIDFn                 func(ctx context.Context, id int) (*entity.Promotion, error)
	getByOrgIDFn              func(ctx context.Context, orgID int) ([]entity.Promotion, error)
	updateFn                  func(ctx context.Context, p *entity.Promotion) error
	deleteFn                  func(ctx context.Context, id int) error
	getApplicableFn           func(ctx context.Context, orgID, clientID int) ([]entity.Promotion, error)
	recordUsageFn             func(ctx context.Context, promotionID, clientID int, promoCodeID *int) error
	createPromoCodeFn         func(ctx context.Context, pc *entity.PromoCode) error
	getPromoCodeByCodeFn      func(ctx context.Context, orgID int, code string) (*entity.PromoCode, error)
	getPromoCodeByIDFn        func(ctx context.Context, id int) (*entity.PromoCode, error)
	listPromoCodesFn          func(ctx context.Context, orgID int) ([]entity.PromoCode, error)
	incrementPromoCodeUsageFn func(ctx context.Context, id int) error
	deactivatePromoCodeFn     func(ctx context.Context, id int) error
}

func (m *mockPromotionsRepo) Create(ctx context.Context, p *entity.Promotion) error {
	return m.createFn(ctx, p)
}
func (m *mockPromotionsRepo) GetByID(ctx context.Context, id int) (*entity.Promotion, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockPromotionsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Promotion, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockPromotionsRepo) Update(ctx context.Context, p *entity.Promotion) error {
	return m.updateFn(ctx, p)
}
func (m *mockPromotionsRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockPromotionsRepo) GetApplicable(ctx context.Context, orgID, clientID int) ([]entity.Promotion, error) {
	return m.getApplicableFn(ctx, orgID, clientID)
}
func (m *mockPromotionsRepo) RecordUsage(ctx context.Context, promotionID, clientID int, promoCodeID *int) error {
	return m.recordUsageFn(ctx, promotionID, clientID, promoCodeID)
}
func (m *mockPromotionsRepo) CreatePromoCode(ctx context.Context, pc *entity.PromoCode) error {
	return m.createPromoCodeFn(ctx, pc)
}
func (m *mockPromotionsRepo) GetPromoCodeByCode(ctx context.Context, orgID int, code string) (*entity.PromoCode, error) {
	return m.getPromoCodeByCodeFn(ctx, orgID, code)
}
func (m *mockPromotionsRepo) GetPromoCodeByID(ctx context.Context, id int) (*entity.PromoCode, error) {
	return m.getPromoCodeByIDFn(ctx, id)
}
func (m *mockPromotionsRepo) ListPromoCodes(ctx context.Context, orgID int) ([]entity.PromoCode, error) {
	return m.listPromoCodesFn(ctx, orgID)
}
func (m *mockPromotionsRepo) IncrementPromoCodeUsage(ctx context.Context, id int) error {
	return m.incrementPromoCodeUsageFn(ctx, id)
}
func (m *mockPromotionsRepo) DeactivatePromoCode(ctx context.Context, id int) error {
	return m.deactivatePromoCodeFn(ctx, id)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreate_SetsOrgAndActive(t *testing.T) {
	repo := &mockPromotionsRepo{
		createFn: func(_ context.Context, p *entity.Promotion) error {
			p.ID = 1
			return nil
		},
	}
	uc := New(repo)

	p, err := uc.Create(context.Background(), 10, &entity.CreatePromotionRequest{
		Name: "Summer Deal",
		Type: "discount",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.OrgID != 10 {
		t.Errorf("expected org_id=10, got %d", p.OrgID)
	}
	if !p.Active {
		t.Error("expected promotion to be active on creation")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockPromotionsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Promotion, error) {
			return nil, fmt.Errorf("promotions.GetByID: %w", sql.ErrNoRows)
		},
	}
	uc := New(repo)

	_, err := uc.GetByID(context.Background(), 99, 1)
	if err != ErrPromotionNotFound {
		t.Errorf("expected ErrPromotionNotFound, got: %v", err)
	}
}

func TestApplyPromoCode_NotFound(t *testing.T) {
	repo := &mockPromotionsRepo{
		getPromoCodeByCodeFn: func(_ context.Context, _ int, _ string) (*entity.PromoCode, error) {
			return nil, fmt.Errorf("promotions.GetPromoCodeByCode: %w", sql.ErrNoRows)
		},
	}
	uc := New(repo)

	_, err := uc.ApplyPromoCode(context.Background(), 1, 1, "INVALID")
	if err != ErrPromoCodeNotFound {
		t.Errorf("expected ErrPromoCodeNotFound, got: %v", err)
	}
}

func TestApplyPromoCode_Expired(t *testing.T) {
	past := time.Now().AddDate(0, 0, -1)
	repo := &mockPromotionsRepo{
		getPromoCodeByCodeFn: func(_ context.Context, _ int, _ string) (*entity.PromoCode, error) {
			return &entity.PromoCode{
				ID:     1,
				OrgID:  1,
				Code:   "EXPIRED",
				Active: true,
				EndsAt: &past,
			}, nil
		},
	}
	uc := New(repo)

	_, err := uc.ApplyPromoCode(context.Background(), 1, 1, "EXPIRED")
	if err != ErrPromoCodeExpired {
		t.Errorf("expected ErrPromoCodeExpired, got: %v", err)
	}
}

func TestApplyPromoCode_LimitReached(t *testing.T) {
	limit := 5
	repo := &mockPromotionsRepo{
		getPromoCodeByCodeFn: func(_ context.Context, _ int, _ string) (*entity.PromoCode, error) {
			return &entity.PromoCode{
				ID:         1,
				OrgID:      1,
				Code:       "LIMITED",
				Active:     true,
				UsageCount: 5,
				UsageLimit: &limit,
			}, nil
		},
	}
	uc := New(repo)

	_, err := uc.ApplyPromoCode(context.Background(), 1, 1, "LIMITED")
	if err != ErrPromoCodeLimitReached {
		t.Errorf("expected ErrPromoCodeLimitReached, got: %v", err)
	}
}

func TestApplyPromoCode_Success(t *testing.T) {
	pct := 10.0
	repo := &mockPromotionsRepo{
		getPromoCodeByCodeFn: func(_ context.Context, _ int, _ string) (*entity.PromoCode, error) {
			return &entity.PromoCode{
				ID:              1,
				OrgID:           1,
				Code:            "SAVE10",
				Active:          true,
				DiscountPercent: &pct,
			}, nil
		},
		incrementPromoCodeUsageFn: func(_ context.Context, _ int) error { return nil },
		recordUsageFn:             func(_ context.Context, _, _ int, _ *int) error { return nil },
	}
	uc := New(repo)

	result, err := uc.ApplyPromoCode(context.Background(), 1, 1, "SAVE10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountPercent == nil || *result.DiscountPercent != 10.0 {
		t.Errorf("expected discount 10%%, got %v", result.DiscountPercent)
	}
}
