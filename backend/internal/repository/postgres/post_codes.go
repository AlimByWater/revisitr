package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type PostCodes struct {
	pg *Module
}

func NewPostCodes(pg *Module) *PostCodes {
	return &PostCodes{pg: pg}
}

func (r *PostCodes) Create(ctx context.Context, pc *entity.PostCode) error {
	err := r.pg.DB().QueryRowContext(ctx, `
		INSERT INTO post_codes (org_id, code, content, created_by_telegram_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		pc.OrgID, pc.Code, pc.Content, pc.CreatedByTelegramID,
	).Scan(&pc.ID, &pc.CreatedAt, &pc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("post_codes.Create: %w", err)
	}
	return nil
}

func (r *PostCodes) GetByCode(ctx context.Context, orgID int, code string) (*entity.PostCode, error) {
	var pc entity.PostCode
	err := r.pg.DB().GetContext(ctx, &pc,
		"SELECT * FROM post_codes WHERE org_id = $1 AND code = $2", orgID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("post_codes.GetByCode: %w", err)
	}
	return &pc, nil
}

func (r *PostCodes) GetByOrgID(ctx context.Context, orgID int) ([]entity.PostCode, error) {
	var codes []entity.PostCode
	err := r.pg.DB().SelectContext(ctx, &codes,
		"SELECT * FROM post_codes WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("post_codes.GetByOrgID: %w", err)
	}
	return codes, nil
}

func (r *PostCodes) Delete(ctx context.Context, orgID int, code string) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM post_codes WHERE org_id = $1 AND code = $2", orgID, code)
	if err != nil {
		return fmt.Errorf("post_codes.Delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostCodes) UpdateContent(ctx context.Context, id int, content entity.PostCodeContent) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"UPDATE post_codes SET content = $1, updated_at = NOW() WHERE id = $2",
		content, id)
	if err != nil {
		return fmt.Errorf("post_codes.UpdateContent: %w", err)
	}
	return nil
}
