package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type Campaigns struct {
	pg *Module
}

func NewCampaigns(pg *Module) *Campaigns {
	return &Campaigns{pg: pg}
}

func (r *Campaigns) Create(ctx context.Context, campaign *entity.Campaign) error {
	query := `
		INSERT INTO campaigns (org_id, bot_id, name, type, status, audience_filter, message, media_url, scheduled_at, stats)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	filterVal, err := campaign.AudienceFilter.Value()
	if err != nil {
		return fmt.Errorf("campaigns.Create filter value: %w", err)
	}

	statsVal, err := campaign.Stats.Value()
	if err != nil {
		return fmt.Errorf("campaigns.Create stats value: %w", err)
	}

	err = r.pg.DB().QueryRowContext(ctx, query,
		campaign.OrgID, campaign.BotID, campaign.Name, campaign.Type, campaign.Status,
		filterVal, campaign.Message, campaign.MediaURL, campaign.ScheduledAt, statsVal,
	).Scan(&campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return fmt.Errorf("campaigns.Create: %w", err)
	}

	return nil
}

func (r *Campaigns) GetByID(ctx context.Context, id int) (*entity.Campaign, error) {
	var campaign entity.Campaign
	err := r.pg.DB().GetContext(ctx, &campaign, "SELECT * FROM campaigns WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("campaigns.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("campaigns.GetByID: %w", err)
	}
	return &campaign, nil
}

func (r *Campaigns) GetByOrgID(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.pg.DB().GetContext(ctx, &total, "SELECT COUNT(*) FROM campaigns WHERE org_id = $1", orgID)
	if err != nil {
		return nil, 0, fmt.Errorf("campaigns.GetByOrgID count: %w", err)
	}

	var campaigns []entity.Campaign
	query := `SELECT * FROM campaigns WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err = r.pg.DB().SelectContext(ctx, &campaigns, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("campaigns.GetByOrgID: %w", err)
	}

	return campaigns, total, nil
}

func (r *Campaigns) Update(ctx context.Context, campaign *entity.Campaign) error {
	filterVal, err := campaign.AudienceFilter.Value()
	if err != nil {
		return fmt.Errorf("campaigns.Update filter value: %w", err)
	}

	query := `
		UPDATE campaigns
		SET name = $1, message = $2, audience_filter = $3, media_url = $4, scheduled_at = $5, updated_at = NOW()
		WHERE id = $6`

	result, err := r.pg.DB().ExecContext(ctx, query,
		campaign.Name, campaign.Message, filterVal, campaign.MediaURL, campaign.ScheduledAt, campaign.ID,
	)
	if err != nil {
		return fmt.Errorf("campaigns.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.Update rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.Update: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Campaigns) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE campaigns SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.pg.DB().ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("campaigns.UpdateStatus: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateStatus rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.UpdateStatus: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Campaigns) UpdateStats(ctx context.Context, id int, stats *entity.CampaignStats) error {
	statsVal, err := stats.Value()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateStats stats value: %w", err)
	}

	query := `UPDATE campaigns SET stats = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.pg.DB().ExecContext(ctx, query, statsVal, id)
	if err != nil {
		return fmt.Errorf("campaigns.UpdateStats: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateStats rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.UpdateStats: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Campaigns) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM campaigns WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("campaigns.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.Delete: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Campaigns) CreateMessagesBatch(ctx context.Context, messages []entity.CampaignMessage) error {
	query := `
		INSERT INTO campaign_messages (campaign_id, client_id, telegram_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	for i := range messages {
		err := r.pg.DB().QueryRowContext(ctx, query,
			messages[i].CampaignID, messages[i].ClientID, messages[i].TelegramID, messages[i].Status,
		).Scan(&messages[i].ID, &messages[i].CreatedAt)
		if err != nil {
			return fmt.Errorf("campaigns.CreateMessagesBatch: %w", err)
		}
	}

	return nil
}

func (r *Campaigns) UpdateMessageStatus(ctx context.Context, id int, status string, errorMsg *string) error {
	var sentAt *time.Time
	if status == "sent" {
		now := time.Now()
		sentAt = &now
	}

	query := `UPDATE campaign_messages SET status = $1, error_message = $2, sent_at = $3 WHERE id = $4`
	result, err := r.pg.DB().ExecContext(ctx, query, status, errorMsg, sentAt, id)
	if err != nil {
		return fmt.Errorf("campaigns.UpdateMessageStatus: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateMessageStatus rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.UpdateMessageStatus: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Campaigns) GetMessagesByCampaignID(ctx context.Context, campaignID, limit, offset int) ([]entity.CampaignMessage, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.pg.DB().GetContext(ctx, &total, "SELECT COUNT(*) FROM campaign_messages WHERE campaign_id = $1", campaignID)
	if err != nil {
		return nil, 0, fmt.Errorf("campaigns.GetMessagesByCampaignID count: %w", err)
	}

	var messages []entity.CampaignMessage
	query := `SELECT * FROM campaign_messages WHERE campaign_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err = r.pg.DB().SelectContext(ctx, &messages, query, campaignID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("campaigns.GetMessagesByCampaignID: %w", err)
	}

	return messages, total, nil
}

func (r *Campaigns) GetMessageStats(ctx context.Context, campaignID int) (*entity.CampaignStats, error) {
	var stats entity.CampaignStats

	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'sent') as sent,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM campaign_messages
		WHERE campaign_id = $1`

	err := r.pg.DB().GetContext(ctx, &stats, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetMessageStats: %w", err)
	}

	return &stats, nil
}
