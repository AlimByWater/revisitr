package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type AdminBot struct {
	pg *Module
}

func NewAdminBot(pg *Module) *AdminBot {
	return &AdminBot{pg: pg}
}

func (r *AdminBot) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.AdminBotLink, error) {
	var link entity.AdminBotLink
	err := r.pg.DB().GetContext(ctx, &link,
		"SELECT * FROM admin_bot_links WHERE telegram_id = $1", telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin_bot.GetByTelegramID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("admin_bot.GetByTelegramID: %w", err)
	}
	return &link, nil
}

func (r *AdminBot) GetByUserID(ctx context.Context, userID int) (*entity.AdminBotLink, error) {
	var link entity.AdminBotLink
	err := r.pg.DB().GetContext(ctx, &link,
		"SELECT * FROM admin_bot_links WHERE user_id = $1", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin_bot.GetByUserID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("admin_bot.GetByUserID: %w", err)
	}
	return &link, nil
}

func (r *AdminBot) CreateLinkCode(ctx context.Context, userID int, orgID int, role string, code string, expiresAt time.Time) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		INSERT INTO admin_bot_links (user_id, org_id, role, link_code, link_code_expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) WHERE telegram_id IS NULL
		DO UPDATE SET link_code = $4, link_code_expires_at = $5`,
		userID, orgID, role, code, expiresAt)
	if err != nil {
		return fmt.Errorf("admin_bot.CreateLinkCode: %w", err)
	}
	return nil
}

func (r *AdminBot) GetByLinkCode(ctx context.Context, code string) (*entity.AdminBotLink, error) {
	var link entity.AdminBotLink
	err := r.pg.DB().GetContext(ctx, &link, `
		SELECT * FROM admin_bot_links
		WHERE link_code = $1 AND link_code_expires_at > now()`, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin_bot.GetByLinkCode: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("admin_bot.GetByLinkCode: %w", err)
	}
	return &link, nil
}

func (r *AdminBot) ActivateLink(ctx context.Context, id int, telegramID int64) error {
	now := time.Now()
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE admin_bot_links
		SET telegram_id = $1, linked_at = $2, link_code = NULL, link_code_expires_at = NULL
		WHERE id = $3`, telegramID, now, id)
	if err != nil {
		return fmt.Errorf("admin_bot.ActivateLink: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("admin_bot.ActivateLink: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *AdminBot) DeleteLink(ctx context.Context, userID int) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM admin_bot_links WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("admin_bot.DeleteLink: %w", err)
	}
	return nil
}
