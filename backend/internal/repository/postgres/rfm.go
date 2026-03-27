package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type RFM struct {
	pg *Module
}

func NewRFM(pg *Module) *RFM {
	return &RFM{pg: pg}
}

func (r *RFM) GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error) {
	var cfg entity.RFMConfig
	err := r.pg.DB().GetContext(ctx, &cfg,
		"SELECT * FROM rfm_configs WHERE org_id = $1", orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // no config yet
		}
		return nil, fmt.Errorf("rfm.GetConfig: %w", err)
	}
	return &cfg, nil
}

func (r *RFM) UpsertConfig(ctx context.Context, cfg *entity.RFMConfig) error {
	query := `
		INSERT INTO rfm_configs (
			org_id, period_days, recalc_interval,
			active_template_type, active_template_key,
			custom_template_name, custom_r_thresholds, custom_f_thresholds
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (org_id) DO UPDATE
		SET period_days          = EXCLUDED.period_days,
		    recalc_interval      = EXCLUDED.recalc_interval,
		    active_template_type = EXCLUDED.active_template_type,
		    active_template_key  = EXCLUDED.active_template_key,
		    custom_template_name = EXCLUDED.custom_template_name,
		    custom_r_thresholds  = EXCLUDED.custom_r_thresholds,
		    custom_f_thresholds  = EXCLUDED.custom_f_thresholds,
		    updated_at           = NOW()
		RETURNING id, created_at, updated_at`

	templateType := cfg.ActiveTemplateType
	if templateType == "" {
		templateType = "standard"
	}
	templateKey := cfg.ActiveTemplateKey
	if templateKey == "" {
		templateKey = "tsr"
	}

	return r.pg.DB().QueryRowContext(ctx, query,
		cfg.OrgID, cfg.PeriodDays, cfg.RecalcInterval,
		templateType, templateKey,
		cfg.CustomTemplateName, cfg.CustomRThresholds, cfg.CustomFThresholds,
	).Scan(&cfg.ID, &cfg.CreatedAt, &cfg.UpdatedAt)
}

func (r *RFM) UpdateCalcStats(ctx context.Context, orgID, clientsProcessed int) error {
	now := time.Now()
	_, err := r.pg.DB().ExecContext(ctx, `
		UPDATE rfm_configs
		SET last_calc_at = $1, clients_processed = $2, updated_at = NOW()
		WHERE org_id = $3`,
		now, clientsProcessed, orgID)
	if err != nil {
		return fmt.Errorf("rfm.UpdateCalcStats: %w", err)
	}
	return nil
}

func (r *RFM) InsertHistory(ctx context.Context, h *entity.RFMHistory) error {
	query := `
		INSERT INTO rfm_history (org_id, segment, client_count, calculated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	return r.pg.DB().QueryRowContext(ctx, query,
		h.OrgID, h.Segment, h.ClientCount, h.CalculatedAt,
	).Scan(&h.ID)
}

func (r *RFM) GetHistory(ctx context.Context, orgID int, from, to time.Time) ([]entity.RFMHistory, error) {
	var history []entity.RFMHistory
	err := r.pg.DB().SelectContext(ctx, &history, `
		SELECT * FROM rfm_history
		WHERE org_id = $1 AND calculated_at BETWEEN $2 AND $3
		ORDER BY calculated_at, segment`,
		orgID, from, to)
	if err != nil {
		return nil, fmt.Errorf("rfm.GetHistory: %w", err)
	}
	return history, nil
}

func (r *RFM) GetSegmentSummary(ctx context.Context, orgID int) ([]entity.RFMSegmentSummary, error) {
	var summaries []entity.RFMSegmentSummary
	err := r.pg.DB().SelectContext(ctx, &summaries, `
		SELECT
			COALESCE(bc.rfm_segment, 'unknown') AS segment,
			COUNT(*) AS client_count,
			COALESCE(AVG(bc.monetary_sum), 0) AS avg_check,
			COALESCE(SUM(bc.monetary_sum), 0) AS total_check
		FROM bot_clients bc
		JOIN bots b ON b.id = bc.bot_id
		WHERE b.org_id = $1 AND bc.rfm_segment IS NOT NULL
		GROUP BY bc.rfm_segment
		ORDER BY client_count DESC`,
		orgID)
	if err != nil {
		return nil, fmt.Errorf("rfm.GetSegmentSummary: %w", err)
	}

	// Calculate percentages
	var total int
	for _, s := range summaries {
		total += s.ClientCount
	}
	for i := range summaries {
		if total > 0 {
			summaries[i].Percentage = float64(summaries[i].ClientCount) / float64(total) * 100
		}
	}

	return summaries, nil
}

// GetSegmentClients returns paginated client list for a specific RFM segment.
func (r *RFM) GetSegmentClients(ctx context.Context, orgID int, segment string, sortCol, order string, limit, offset int) ([]entity.SegmentClientRow, int, error) {
	// Whitelist sort columns to prevent SQL injection
	allowedSort := map[string]string{
		"r_score":          "bc.r_score",
		"f_score":          "bc.f_score",
		"m_score":          "bc.m_score",
		"last_visit_date":  "bc.last_visit_date",
		"frequency_count":  "bc.frequency_count",
		"monetary_sum":     "bc.monetary_sum",
	}
	col, ok := allowedSort[sortCol]
	if !ok {
		col = "bc.monetary_sum"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}

	// Count total
	var total int
	err := r.pg.DB().GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM bot_clients bc
		JOIN bots b ON b.id = bc.bot_id
		WHERE b.org_id = $1 AND bc.rfm_segment = $2`,
		orgID, segment)
	if err != nil {
		return nil, 0, fmt.Errorf("rfm.GetSegmentClients count: %w", err)
	}

	var rows []entity.SegmentClientRow
	query := fmt.Sprintf(`
		SELECT
			bc.id, bc.first_name, bc.last_name, bc.phone,
			bc.r_score, bc.f_score, bc.m_score,
			bc.recency_days, bc.frequency_count, bc.monetary_sum,
			bc.last_visit_date, bc.total_visits_lifetime
		FROM bot_clients bc
		JOIN bots b ON b.id = bc.bot_id
		WHERE b.org_id = $1 AND bc.rfm_segment = $2
		ORDER BY %s %s NULLS LAST
		LIMIT $3 OFFSET $4`, col, dir)

	err = r.pg.DB().SelectContext(ctx, &rows, query, orgID, segment, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("rfm.GetSegmentClients: %w", err)
	}

	return rows, total, nil
}
