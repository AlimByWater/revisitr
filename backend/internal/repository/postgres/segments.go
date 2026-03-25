package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Segments struct {
	pg *Module
}

func NewSegments(pg *Module) *Segments {
	return &Segments{pg: pg}
}

func (r *Segments) Create(ctx context.Context, seg *entity.Segment) error {
	filterVal, err := seg.Filter.Value()
	if err != nil {
		return fmt.Errorf("segments.Create filter value: %w", err)
	}

	query := `
		INSERT INTO segments (org_id, name, type, filter, auto_assign)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		seg.OrgID, seg.Name, seg.Type, filterVal, seg.AutoAssign,
	).Scan(&seg.ID, &seg.CreatedAt, &seg.UpdatedAt)
}

func (r *Segments) GetByID(ctx context.Context, id int) (*entity.Segment, error) {
	var seg entity.Segment
	err := r.pg.DB().GetContext(ctx, &seg, "SELECT * FROM segments WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("segments.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("segments.GetByID: %w", err)
	}
	return &seg, nil
}

func (r *Segments) GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error) {
	var segs []entity.Segment
	err := r.pg.DB().SelectContext(ctx, &segs,
		"SELECT * FROM segments WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("segments.GetByOrgID: %w", err)
	}
	return segs, nil
}

func (r *Segments) Update(ctx context.Context, seg *entity.Segment) error {
	filterVal, err := seg.Filter.Value()
	if err != nil {
		return fmt.Errorf("segments.Update filter value: %w", err)
	}

	query := `
		UPDATE segments
		SET name = $1, filter = $2, auto_assign = $3, updated_at = NOW()
		WHERE id = $4`

	result, err := r.pg.DB().ExecContext(ctx, query,
		seg.Name, filterVal, seg.AutoAssign, seg.ID)
	if err != nil {
		return fmt.Errorf("segments.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("segments.Update rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("segments.Update: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Segments) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM segments WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("segments.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("segments.Delete rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("segments.Delete: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Segments) GetClients(ctx context.Context, segmentID, limit, offset int) ([]entity.BotClient, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.pg.DB().GetContext(ctx, &total,
		"SELECT COUNT(*) FROM segment_clients WHERE segment_id = $1", segmentID)
	if err != nil {
		return nil, 0, fmt.Errorf("segments.GetClients count: %w", err)
	}

	query := `
		SELECT bc.*
		FROM bot_clients bc
		JOIN segment_clients sc ON sc.client_id = bc.id
		WHERE sc.segment_id = $1
		ORDER BY sc.assigned_at DESC
		LIMIT $2 OFFSET $3`

	var clients []entity.BotClient
	if err := r.pg.DB().SelectContext(ctx, &clients, query, segmentID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("segments.GetClients: %w", err)
	}

	return clients, total, nil
}

func (r *Segments) SyncClients(ctx context.Context, segmentID int, clientIDs []int) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("segments.SyncClients begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM segment_clients WHERE segment_id = $1", segmentID); err != nil {
		return fmt.Errorf("segments.SyncClients delete: %w", err)
	}

	for _, cid := range clientIDs {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO segment_clients (segment_id, client_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			segmentID, cid); err != nil {
			return fmt.Errorf("segments.SyncClients insert: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Segments) CountByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error) {
	query := `
		SELECT COUNT(DISTINCT bc.id)
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1`

	args := []interface{}{orgID}
	idx := 2

	if f.Gender != nil {
		query += fmt.Sprintf(" AND bc.gender = $%d", idx)
		args = append(args, *f.Gender)
		idx++
	}
	if f.Tags != nil && len(f.Tags) > 0 {
		for _, tag := range f.Tags {
			query += fmt.Sprintf(" AND bc.tags @> $%d::jsonb", idx)
			args = append(args, `["`+tag+`"]`)
			idx++
		}
	}

	var count int
	if err := r.pg.DB().GetContext(ctx, &count, query, args...); err != nil {
		return 0, fmt.Errorf("segments.CountByFilter: %w", err)
	}
	return count, nil
}

// ── Segment Rules ────────────────────────────────────────────────────────────

func (r *Segments) CreateRule(ctx context.Context, segmentID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error) {
	var rule entity.SegmentRule
	err := r.pg.DB().GetContext(ctx, &rule, `
		INSERT INTO segment_rules (segment_id, field, operator, value)
		VALUES ($1, $2, $3, $4)
		RETURNING *`,
		segmentID, req.Field, req.Operator, req.Value)
	if err != nil {
		return nil, fmt.Errorf("segments.CreateRule: %w", err)
	}
	return &rule, nil
}

func (r *Segments) GetRules(ctx context.Context, segmentID int) ([]entity.SegmentRule, error) {
	var rules []entity.SegmentRule
	err := r.pg.DB().SelectContext(ctx, &rules,
		"SELECT * FROM segment_rules WHERE segment_id = $1 ORDER BY id", segmentID)
	if err != nil {
		return nil, fmt.Errorf("segments.GetRules: %w", err)
	}
	return rules, nil
}

func (r *Segments) DeleteRule(ctx context.Context, ruleID int) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM segment_rules WHERE id = $1", ruleID)
	if err != nil {
		return fmt.Errorf("segments.DeleteRule: %w", err)
	}
	return nil
}

func (r *Segments) DeleteRulesBySegment(ctx context.Context, segmentID int) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM segment_rules WHERE segment_id = $1", segmentID)
	if err != nil {
		return fmt.Errorf("segments.DeleteRulesBySegment: %w", err)
	}
	return nil
}

// ── Client Predictions ───────────────────────────────────────────────────────

func (r *Segments) UpsertPrediction(ctx context.Context, pred *entity.ClientPrediction) error {
	err := r.pg.DB().GetContext(ctx, pred, `
		INSERT INTO client_predictions (org_id, client_id, churn_risk, upsell_score, predicted_value, factors)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (client_id) DO UPDATE SET
			churn_risk = EXCLUDED.churn_risk,
			upsell_score = EXCLUDED.upsell_score,
			predicted_value = EXCLUDED.predicted_value,
			factors = EXCLUDED.factors,
			computed_at = now()
		RETURNING *`,
		pred.OrgID, pred.ClientID, pred.ChurnRisk, pred.UpsellScore, pred.PredictedValue, pred.Factors)
	if err != nil {
		return fmt.Errorf("segments.UpsertPrediction: %w", err)
	}
	return nil
}

func (r *Segments) GetPredictions(ctx context.Context, orgID int, limit, offset int) ([]entity.ClientPrediction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var preds []entity.ClientPrediction
	err := r.pg.DB().SelectContext(ctx, &preds, `
		SELECT * FROM client_predictions
		WHERE org_id = $1
		ORDER BY churn_risk DESC
		LIMIT $2 OFFSET $3`,
		orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("segments.GetPredictions: %w", err)
	}
	return preds, nil
}

func (r *Segments) GetPredictionByClient(ctx context.Context, clientID int) (*entity.ClientPrediction, error) {
	var pred entity.ClientPrediction
	err := r.pg.DB().GetContext(ctx, &pred,
		"SELECT * FROM client_predictions WHERE client_id = $1", clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("segments.GetPredictionByClient: %w", err)
	}
	return &pred, nil
}

func (r *Segments) GetHighChurnClients(ctx context.Context, orgID int, threshold float32) ([]entity.ClientPrediction, error) {
	var preds []entity.ClientPrediction
	err := r.pg.DB().SelectContext(ctx, &preds, `
		SELECT * FROM client_predictions
		WHERE org_id = $1 AND churn_risk >= $2
		ORDER BY churn_risk DESC`,
		orgID, threshold)
	if err != nil {
		return nil, fmt.Errorf("segments.GetHighChurnClients: %w", err)
	}
	return preds, nil
}

func (r *Segments) GetPredictionSummary(ctx context.Context, orgID int) (*entity.PredictionSummary, error) {
	var summary entity.PredictionSummary
	err := r.pg.DB().GetContext(ctx, &summary, `
		SELECT
			COUNT(*) FILTER (WHERE churn_risk >= 0.7) AS high_churn_count,
			COALESCE(AVG(churn_risk), 0) AS avg_churn_risk,
			COUNT(*) FILTER (WHERE upsell_score >= 0.7) AS high_upsell_count,
			COUNT(*) AS total_predicted
		FROM client_predictions WHERE org_id = $1`, orgID)
	if err != nil {
		return nil, fmt.Errorf("segments.GetPredictionSummary: %w", err)
	}
	return &summary, nil
}
