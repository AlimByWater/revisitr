package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type Integrations struct {
	pg *Module
}

func NewIntegrations(pg *Module) *Integrations {
	return &Integrations{pg: pg}
}

func (r *Integrations) Create(ctx context.Context, intg *entity.Integration) error {
	cfgVal, err := intg.Config.Value()
	if err != nil {
		return fmt.Errorf("integrations.Create config value: %w", err)
	}

	query := `
		INSERT INTO integrations (org_id, type, config, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		intg.OrgID, intg.Type, cfgVal, intg.Status,
	).Scan(&intg.ID, &intg.CreatedAt, &intg.UpdatedAt)
}

func (r *Integrations) GetByID(ctx context.Context, id int) (*entity.Integration, error) {
	var intg entity.Integration
	err := r.pg.DB().GetContext(ctx, &intg, "SELECT * FROM integrations WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("integrations.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("integrations.GetByID: %w", err)
	}
	return &intg, nil
}

func (r *Integrations) GetByOrgID(ctx context.Context, orgID int) ([]entity.Integration, error) {
	var intgs []entity.Integration
	err := r.pg.DB().SelectContext(ctx, &intgs,
		"SELECT * FROM integrations WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("integrations.GetByOrgID: %w", err)
	}
	return intgs, nil
}

func (r *Integrations) Update(ctx context.Context, intg *entity.Integration) error {
	cfgVal, err := intg.Config.Value()
	if err != nil {
		return fmt.Errorf("integrations.Update config value: %w", err)
	}

	query := `
		UPDATE integrations
		SET config = $1, status = $2, updated_at = NOW()
		WHERE id = $3`

	result, err := r.pg.DB().ExecContext(ctx, query, cfgVal, intg.Status, intg.ID)
	if err != nil {
		return fmt.Errorf("integrations.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("integrations.Update rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("integrations.Update: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Integrations) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM integrations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("integrations.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("integrations.Delete rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("integrations.Delete: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Integrations) GetActive(ctx context.Context) ([]entity.Integration, error) {
	var intgs []entity.Integration
	err := r.pg.DB().SelectContext(ctx, &intgs,
		"SELECT * FROM integrations WHERE status = 'active' ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("integrations.GetActive: %w", err)
	}
	return intgs, nil
}

func (r *Integrations) GetOrdersByIntegration(ctx context.Context, integrationID, limit, offset int) ([]entity.ExternalOrder, int, error) {
	var total int
	err := r.pg.DB().GetContext(ctx, &total,
		"SELECT COUNT(*) FROM external_orders WHERE integration_id = $1", integrationID)
	if err != nil {
		return nil, 0, fmt.Errorf("integrations.GetOrdersByIntegration count: %w", err)
	}

	var orders []entity.ExternalOrder
	err = r.pg.DB().SelectContext(ctx, &orders,
		"SELECT * FROM external_orders WHERE integration_id = $1 ORDER BY ordered_at DESC NULLS LAST LIMIT $2 OFFSET $3",
		integrationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("integrations.GetOrdersByIntegration: %w", err)
	}
	return orders, total, nil
}

func (r *Integrations) GetSyncStats(ctx context.Context, integrationID int) (*entity.IntegrationStats, error) {
	stats := &entity.IntegrationStats{}
	err := r.pg.DB().GetContext(ctx, stats, `
		SELECT
			COUNT(*) as total_orders,
			COALESCE(SUM(total), 0) as total_revenue,
			COUNT(DISTINCT client_id) FILTER (WHERE client_id IS NOT NULL) as matched_clients,
			COUNT(*) FILTER (WHERE client_id IS NULL) as unmatched_orders
		FROM external_orders
		WHERE integration_id = $1`, integrationID)
	if err != nil {
		return nil, fmt.Errorf("integrations.GetSyncStats: %w", err)
	}
	return stats, nil
}

func (r *Integrations) UpdateLastSync(ctx context.Context, id int, status string) error {
	now := time.Now()
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE integrations SET last_sync_at = $1, status = $2, updated_at = NOW() WHERE id = $3",
		now, status, id)
	if err != nil {
		return fmt.Errorf("integrations.UpdateLastSync: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("integrations.UpdateLastSync rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("integrations.UpdateLastSync: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Integrations) UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error {
	itemsVal, err := order.Items.Value()
	if err != nil {
		return fmt.Errorf("integrations.UpsertOrder items value: %w", err)
	}

	query := `
		INSERT INTO external_orders (integration_id, external_id, client_id, items, total, ordered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (integration_id, external_id) DO UPDATE
		SET client_id = EXCLUDED.client_id,
		    items     = EXCLUDED.items,
		    total     = EXCLUDED.total,
		    ordered_at = EXCLUDED.ordered_at,
		    synced_at  = NOW()
		RETURNING id, synced_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		order.IntegrationID, order.ExternalID, order.ClientID, itemsVal, order.Total, order.OrderedAt,
	).Scan(&order.ID, &order.SyncedAt)
}
