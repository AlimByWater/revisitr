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

func (r *Campaigns) UpdateContent(ctx context.Context, id int, content entity.MessageContent) error {
	query := `UPDATE campaigns SET content = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.pg.DB().ExecContext(ctx, query, content, id)
	if err != nil {
		return fmt.Errorf("campaigns.UpdateContent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateContent rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.UpdateContent: %w", sql.ErrNoRows)
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

func (r *Campaigns) CreateClick(ctx context.Context, click *entity.CampaignClick) error {
	query := `
		INSERT INTO campaign_clicks (campaign_id, client_id, button_idx, url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, clicked_at`

	err := r.pg.DB().QueryRowContext(ctx, query,
		click.CampaignID, click.ClientID, click.ButtonIdx, click.URL,
	).Scan(&click.ID, &click.ClickedAt)
	if err != nil {
		return fmt.Errorf("campaigns.CreateClick: %w", err)
	}

	return nil
}

func (r *Campaigns) GetClicksByCampaign(ctx context.Context, campaignID int) ([]entity.CampaignClick, error) {
	var clicks []entity.CampaignClick
	query := `SELECT * FROM campaign_clicks WHERE campaign_id = $1 ORDER BY clicked_at DESC`
	err := r.pg.DB().SelectContext(ctx, &clicks, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetClicksByCampaign: %w", err)
	}

	return clicks, nil
}

func (r *Campaigns) GetCampaignAnalytics(ctx context.Context, campaignID int) (*entity.CampaignAnalyticsDetail, error) {
	var msgStats struct {
		Total  int `db:"total"`
		Sent   int `db:"sent"`
		Failed int `db:"failed"`
	}

	msgQuery := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'sent') as sent,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM campaign_messages
		WHERE campaign_id = $1`

	err := r.pg.DB().GetContext(ctx, &msgStats, msgQuery, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetCampaignAnalytics messages: %w", err)
	}

	var clicked int
	clickQuery := `SELECT COUNT(DISTINCT client_id) FROM campaign_clicks WHERE campaign_id = $1`
	err = r.pg.DB().GetContext(ctx, &clicked, clickQuery, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetCampaignAnalytics clicks: %w", err)
	}

	var clickRate float64
	if msgStats.Sent > 0 {
		clickRate = float64(clicked) / float64(msgStats.Sent)
	}

	return &entity.CampaignAnalyticsDetail{
		Total:     msgStats.Total,
		Sent:      msgStats.Sent,
		Failed:    msgStats.Failed,
		Clicked:   clicked,
		ClickRate: clickRate,
	}, nil
}

func (r *Campaigns) GetScheduledCampaigns(ctx context.Context, before time.Time) ([]entity.Campaign, error) {
	var campaigns []entity.Campaign
	query := `SELECT * FROM campaigns WHERE status = 'scheduled' AND scheduled_at <= $1`
	err := r.pg.DB().SelectContext(ctx, &campaigns, query, before)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetScheduledCampaigns: %w", err)
	}

	return campaigns, nil
}

// ── A/B Variants ─────────────────────────────────────────────────────────────

func (r *Campaigns) CreateVariant(ctx context.Context, v *entity.CampaignVariant) error {
	buttonsVal, err := v.Buttons.Value()
	if err != nil {
		return fmt.Errorf("campaigns.CreateVariant buttons value: %w", err)
	}

	statsVal, err := v.Stats.Value()
	if err != nil {
		return fmt.Errorf("campaigns.CreateVariant stats value: %w", err)
	}

	query := `
		INSERT INTO campaign_variants (campaign_id, name, audience_pct, message, media_url, buttons, stats)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err = r.pg.DB().QueryRowContext(ctx, query,
		v.CampaignID, v.Name, v.AudiencePct, v.Message, v.MediaURL, buttonsVal, statsVal,
	).Scan(&v.ID, &v.CreatedAt)
	if err != nil {
		return fmt.Errorf("campaigns.CreateVariant: %w", err)
	}

	return nil
}

func (r *Campaigns) GetVariantsByCampaignID(ctx context.Context, campaignID int) ([]entity.CampaignVariant, error) {
	var variants []entity.CampaignVariant
	query := `SELECT * FROM campaign_variants WHERE campaign_id = $1 ORDER BY id`
	err := r.pg.DB().SelectContext(ctx, &variants, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetVariantsByCampaignID: %w", err)
	}

	return variants, nil
}

func (r *Campaigns) UpdateVariantStats(ctx context.Context, id int, stats *entity.CampaignStats) error {
	statsVal, err := stats.Value()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateVariantStats stats value: %w", err)
	}

	query := `UPDATE campaign_variants SET stats = $1 WHERE id = $2`
	_, err = r.pg.DB().ExecContext(ctx, query, statsVal, id)
	if err != nil {
		return fmt.Errorf("campaigns.UpdateVariantStats: %w", err)
	}

	return nil
}

func (r *Campaigns) SetVariantWinner(ctx context.Context, campaignID, variantID int) error {
	// Reset all variants for this campaign, then set winner
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("campaigns.SetVariantWinner begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE campaign_variants SET is_winner = false WHERE campaign_id = $1", campaignID)
	if err != nil {
		return fmt.Errorf("campaigns.SetVariantWinner reset: %w", err)
	}

	result, err := tx.ExecContext(ctx, "UPDATE campaign_variants SET is_winner = true WHERE id = $1 AND campaign_id = $2", variantID, campaignID)
	if err != nil {
		return fmt.Errorf("campaigns.SetVariantWinner set: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.SetVariantWinner rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.SetVariantWinner: %w", sql.ErrNoRows)
	}

	return tx.Commit()
}

func (r *Campaigns) DeleteVariantsByCampaignID(ctx context.Context, campaignID int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM campaign_variants WHERE campaign_id = $1", campaignID)
	if err != nil {
		return fmt.Errorf("campaigns.DeleteVariantsByCampaignID: %w", err)
	}

	return nil
}

func (r *Campaigns) GetVariantAnalytics(ctx context.Context, variantID int) (*entity.VariantResult, error) {
	var v entity.CampaignVariant
	err := r.pg.DB().GetContext(ctx, &v, "SELECT * FROM campaign_variants WHERE id = $1", variantID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetVariantAnalytics variant: %w", err)
	}

	var clicked int
	clickQuery := `
		SELECT COUNT(DISTINCT cm.client_id)
		FROM campaign_clicks cc
		JOIN campaign_messages cm ON cm.campaign_id = cc.campaign_id AND cm.client_id = cc.client_id
		WHERE cm.variant_id = $1`
	err = r.pg.DB().GetContext(ctx, &clicked, clickQuery, variantID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetVariantAnalytics clicks: %w", err)
	}

	var clickRate float64
	if v.Stats.Sent > 0 {
		clickRate = float64(clicked) / float64(v.Stats.Sent)
	}

	return &entity.VariantResult{
		ID:          v.ID,
		Name:        v.Name,
		AudiencePct: v.AudiencePct,
		Total:       v.Stats.Total,
		Sent:        v.Stats.Sent,
		Failed:      v.Stats.Failed,
		Clicked:     clicked,
		ClickRate:   clickRate,
		IsWinner:    v.IsWinner,
	}, nil
}

// ── Campaign Templates ───────────────────────────────────────────────────────

func (r *Campaigns) CreateTemplate(ctx context.Context, t *entity.CampaignTemplate) error {
	buttonsVal, err := t.Buttons.Value()
	if err != nil {
		return fmt.Errorf("campaigns.CreateTemplate buttons value: %w", err)
	}

	filterVal, err := t.AudienceFilter.Value()
	if err != nil {
		return fmt.Errorf("campaigns.CreateTemplate filter value: %w", err)
	}

	query := `
		INSERT INTO campaign_templates (org_id, name, category, description, message, media_url, buttons, audience_filter, tracking_mode)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, is_system, created_at, updated_at`

	err = r.pg.DB().QueryRowContext(ctx, query,
		t.OrgID, t.Name, t.Category, t.Description, t.Message, t.MediaURL,
		buttonsVal, filterVal, t.TrackingMode,
	).Scan(&t.ID, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("campaigns.CreateTemplate: %w", err)
	}

	return nil
}

func (r *Campaigns) GetTemplates(ctx context.Context, orgID int) ([]entity.CampaignTemplate, error) {
	var templates []entity.CampaignTemplate
	query := `
		SELECT * FROM campaign_templates
		WHERE org_id = $1 OR org_id IS NULL
		ORDER BY is_system DESC, created_at DESC`
	err := r.pg.DB().SelectContext(ctx, &templates, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("campaigns.GetTemplates: %w", err)
	}

	return templates, nil
}

func (r *Campaigns) GetTemplateByID(ctx context.Context, id int) (*entity.CampaignTemplate, error) {
	var t entity.CampaignTemplate
	err := r.pg.DB().GetContext(ctx, &t, "SELECT * FROM campaign_templates WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("campaigns.GetTemplateByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("campaigns.GetTemplateByID: %w", err)
	}

	return &t, nil
}

func (r *Campaigns) UpdateTemplate(ctx context.Context, t *entity.CampaignTemplate) error {
	buttonsVal, err := t.Buttons.Value()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateTemplate buttons value: %w", err)
	}

	filterVal, err := t.AudienceFilter.Value()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateTemplate filter value: %w", err)
	}

	query := `
		UPDATE campaign_templates
		SET name = $1, category = $2, description = $3, message = $4, media_url = $5,
		    buttons = $6, audience_filter = $7, tracking_mode = $8, updated_at = NOW()
		WHERE id = $9 AND is_system = false`

	result, err := r.pg.DB().ExecContext(ctx, query,
		t.Name, t.Category, t.Description, t.Message, t.MediaURL,
		buttonsVal, filterVal, t.TrackingMode, t.ID,
	)
	if err != nil {
		return fmt.Errorf("campaigns.UpdateTemplate: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.UpdateTemplate rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.UpdateTemplate: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Campaigns) DeleteTemplate(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM campaign_templates WHERE id = $1 AND is_system = false", id)
	if err != nil {
		return fmt.Errorf("campaigns.DeleteTemplate: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("campaigns.DeleteTemplate rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("campaigns.DeleteTemplate: %w", sql.ErrNoRows)
	}

	return nil
}
