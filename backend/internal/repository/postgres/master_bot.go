package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type MasterBot struct {
	pg *Module
}

func NewMasterBot(pg *Module) *MasterBot {
	return &MasterBot{pg: pg}
}

func (r *MasterBot) CreateLink(ctx context.Context, link *entity.MasterBotLink) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		INSERT INTO master_bot_links (org_id, telegram_user_id, telegram_username)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_id, telegram_user_id)
		DO UPDATE SET telegram_username = $3, is_active = true, updated_at = NOW()`,
		link.OrgID, link.TelegramUserID, link.TelegramUsername)
	if err != nil {
		return fmt.Errorf("master_bot.CreateLink: %w", err)
	}
	return nil
}

func (r *MasterBot) GetLinkByTelegramID(ctx context.Context, telegramUserID int64) (*entity.MasterBotLink, error) {
	var link entity.MasterBotLink
	err := r.pg.DB().GetContext(ctx, &link,
		"SELECT * FROM master_bot_links WHERE telegram_user_id = $1 AND is_active = true",
		telegramUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("master_bot.GetLinkByTelegramID: %w", err)
	}
	return &link, nil
}

func (r *MasterBot) GetLinkByOrgID(ctx context.Context, orgID int) ([]entity.MasterBotLink, error) {
	var links []entity.MasterBotLink
	err := r.pg.DB().SelectContext(ctx, &links,
		"SELECT * FROM master_bot_links WHERE org_id = $1 AND is_active = true", orgID)
	if err != nil {
		return nil, fmt.Errorf("master_bot.GetLinkByOrgID: %w", err)
	}
	return links, nil
}

func (r *MasterBot) DeactivateLink(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"UPDATE master_bot_links SET is_active = false, updated_at = $1 WHERE id = $2",
		time.Now(), id)
	if err != nil {
		return fmt.Errorf("master_bot.DeactivateLink: %w", err)
	}
	return nil
}
