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
