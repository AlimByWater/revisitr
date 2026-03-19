package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type Promotions struct {
	pg *Module
}

func NewPromotions(pg *Module) *Promotions {
	return &Promotions{pg: pg}
}

func (r *Promotions) Create(ctx context.Context, p *entity.Promotion) error {
	condVal, err := p.Conditions.Value()
	if err != nil {
		return fmt.Errorf("promotions.Create conditions value: %w", err)
	}
	resultVal, err := p.Result.Value()
	if err != nil {
		return fmt.Errorf("promotions.Create result value: %w", err)
	}

	query := `
		INSERT INTO promotions (org_id, name, type, conditions, result, starts_at, ends_at, usage_limit, combinable, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		p.OrgID, p.Name, p.Type, condVal, resultVal,
		p.StartsAt, p.EndsAt, p.UsageLimit, p.Combinable, p.Active,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *Promotions) GetByID(ctx context.Context, id int) (*entity.Promotion, error) {
	var p entity.Promotion
	err := r.pg.DB().GetContext(ctx, &p, "SELECT * FROM promotions WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("promotions.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("promotions.GetByID: %w", err)
	}
	return &p, nil
}

func (r *Promotions) GetByOrgID(ctx context.Context, orgID int) ([]entity.Promotion, error) {
	var promotions []entity.Promotion
	err := r.pg.DB().SelectContext(ctx, &promotions,
		"SELECT * FROM promotions WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("promotions.GetByOrgID: %w", err)
	}
	return promotions, nil
}

func (r *Promotions) Update(ctx context.Context, p *entity.Promotion) error {
	condVal, err := p.Conditions.Value()
	if err != nil {
		return fmt.Errorf("promotions.Update conditions value: %w", err)
	}
	resultVal, err := p.Result.Value()
	if err != nil {
		return fmt.Errorf("promotions.Update result value: %w", err)
	}

	query := `
		UPDATE promotions
		SET name = $1, conditions = $2, result = $3, starts_at = $4, ends_at = $5,
		    usage_limit = $6, combinable = $7, active = $8
		WHERE id = $9`

	result, err := r.pg.DB().ExecContext(ctx, query,
		p.Name, condVal, resultVal, p.StartsAt, p.EndsAt,
		p.UsageLimit, p.Combinable, p.Active, p.ID)
	if err != nil {
		return fmt.Errorf("promotions.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("promotions.Update rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("promotions.Update: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Promotions) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM promotions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("promotions.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("promotions.Delete rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("promotions.Delete: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Promotions) GetApplicable(ctx context.Context, orgID, clientID int) ([]entity.Promotion, error) {
	now := time.Now()
	query := `
		SELECT p.*
		FROM promotions p
		WHERE p.org_id = $1
		  AND p.active = true
		  AND (p.starts_at IS NULL OR p.starts_at <= $2)
		  AND (p.ends_at IS NULL OR p.ends_at >= $2)
		ORDER BY p.created_at DESC`

	var promotions []entity.Promotion
	if err := r.pg.DB().SelectContext(ctx, &promotions, query, orgID, now); err != nil {
		return nil, fmt.Errorf("promotions.GetApplicable: %w", err)
	}
	return promotions, nil
}

func (r *Promotions) RecordUsage(ctx context.Context, promotionID, clientID int, promoCodeID *int) error {
	_, err := r.pg.DB().ExecContext(ctx,
		`INSERT INTO promotion_usages (promotion_id, client_id, promo_code_id) VALUES ($1, $2, $3)`,
		promotionID, clientID, promoCodeID)
	if err != nil {
		return fmt.Errorf("promotions.RecordUsage: %w", err)
	}
	return nil
}

func (r *Promotions) CreatePromoCode(ctx context.Context, pc *entity.PromoCode) error {
	condVal, err := pc.Conditions.Value()
	if err != nil {
		return fmt.Errorf("promotions.CreatePromoCode conditions value: %w", err)
	}

	query := `
		INSERT INTO promo_codes (org_id, promotion_id, code, discount_percent, bonus_amount,
		                         starts_at, ends_at, conditions, usage_limit, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, usage_count, created_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		pc.OrgID, pc.PromotionID, pc.Code, pc.DiscountPercent, pc.BonusAmount,
		pc.StartsAt, pc.EndsAt, condVal, pc.UsageLimit, pc.Active,
	).Scan(&pc.ID, &pc.UsageCount, &pc.CreatedAt)
}

func (r *Promotions) GetPromoCodeByCode(ctx context.Context, orgID int, code string) (*entity.PromoCode, error) {
	var pc entity.PromoCode
	err := r.pg.DB().GetContext(ctx, &pc,
		"SELECT * FROM promo_codes WHERE org_id = $1 AND code = $2", orgID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("promotions.GetPromoCodeByCode: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("promotions.GetPromoCodeByCode: %w", err)
	}
	return &pc, nil
}

func (r *Promotions) GetPromoCodeByID(ctx context.Context, id int) (*entity.PromoCode, error) {
	var pc entity.PromoCode
	err := r.pg.DB().GetContext(ctx, &pc, "SELECT * FROM promo_codes WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("promotions.GetPromoCodeByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("promotions.GetPromoCodeByID: %w", err)
	}
	return &pc, nil
}

func (r *Promotions) ListPromoCodes(ctx context.Context, orgID int) ([]entity.PromoCode, error) {
	var codes []entity.PromoCode
	err := r.pg.DB().SelectContext(ctx, &codes,
		"SELECT * FROM promo_codes WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("promotions.ListPromoCodes: %w", err)
	}
	return codes, nil
}

func (r *Promotions) IncrementPromoCodeUsage(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE promo_codes SET usage_count = usage_count + 1 WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("promotions.IncrementPromoCodeUsage: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("promotions.IncrementPromoCodeUsage rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("promotions.IncrementPromoCodeUsage: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Promotions) DeactivatePromoCode(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE promo_codes SET active = false WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("promotions.DeactivatePromoCode: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("promotions.DeactivatePromoCode rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("promotions.DeactivatePromoCode: %w", sql.ErrNoRows)
	}
	return nil
}
