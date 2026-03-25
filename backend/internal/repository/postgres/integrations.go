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

func (r *Integrations) UpsertAggregate(ctx context.Context, agg *entity.IntegrationAggregate) error {
	query := `
		INSERT INTO integration_aggregates (integration_id, date, revenue, avg_check, tx_count, guest_count, matched_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (integration_id, date) DO UPDATE
		SET revenue       = EXCLUDED.revenue,
		    avg_check      = EXCLUDED.avg_check,
		    tx_count       = EXCLUDED.tx_count,
		    guest_count    = EXCLUDED.guest_count,
		    matched_count  = EXCLUDED.matched_count,
		    synced_at      = NOW()
		RETURNING id, synced_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		agg.IntegrationID, agg.Date, agg.Revenue, agg.AvgCheck, agg.TxCount, agg.GuestCount, agg.MatchedCount,
	).Scan(&agg.ID, &agg.SyncedAt)
}

func (r *Integrations) GetAggregates(ctx context.Context, integrationID int, from, to time.Time) ([]entity.IntegrationAggregate, error) {
	var aggs []entity.IntegrationAggregate
	err := r.pg.DB().SelectContext(ctx, &aggs,
		"SELECT * FROM integration_aggregates WHERE integration_id = $1 AND date BETWEEN $2 AND $3 ORDER BY date",
		integrationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("integrations.GetAggregates: %w", err)
	}
	return aggs, nil
}

func (r *Integrations) GetDashboardAggregates(ctx context.Context, orgID int, from, to time.Time) (*entity.DashboardAggregates, error) {
	result := &entity.DashboardAggregates{}
	query := `
		SELECT
			COALESCE(SUM(ia.revenue), 0) AS revenue,
			CASE WHEN SUM(ia.tx_count) > 0
				THEN SUM(ia.revenue) / SUM(ia.tx_count)
				ELSE 0
			END AS avg_check,
			COALESCE(SUM(ia.tx_count), 0) AS tx_count,
			COALESCE((
				SELECT CASE WHEN SUM(ia2.tx_count) > 0
					THEN SUM(ia2.revenue) / SUM(ia2.tx_count)
					ELSE 0
				END
				FROM integration_aggregates ia2
				JOIN integrations i2 ON i2.id = ia2.integration_id
				WHERE i2.org_id = $1 AND ia2.date BETWEEN $2 AND $3
				AND ia2.matched_count > 0
			), 0) AS loyalty_avg,
			COALESCE((
				SELECT CASE WHEN SUM(ia3.tx_count - ia3.matched_count) > 0
					THEN SUM(ia3.revenue * (1.0 - CAST(ia3.matched_count AS FLOAT) / NULLIF(ia3.tx_count, 0))) / SUM(ia3.tx_count - ia3.matched_count)
					ELSE 0
				END
				FROM integration_aggregates ia3
				JOIN integrations i3 ON i3.id = ia3.integration_id
				WHERE i3.org_id = $1 AND ia3.date BETWEEN $2 AND $3
				AND ia3.tx_count > ia3.matched_count
			), 0) AS non_loyalty_avg
		FROM integration_aggregates ia
		JOIN integrations i ON i.id = ia.integration_id
		WHERE i.org_id = $1 AND ia.date BETWEEN $2 AND $3`

	err := r.pg.DB().GetContext(ctx, result, query, orgID, from, to)
	if err != nil {
		return nil, fmt.Errorf("integrations.GetDashboardAggregates: %w", err)
	}
	return result, nil
}

func (r *Integrations) UpsertClientMap(ctx context.Context, mapping *entity.IntegrationClientMap) error {
	query := `
		INSERT INTO integration_client_map (integration_id, external_phone, client_id, matched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (integration_id, external_phone) DO UPDATE
		SET client_id   = EXCLUDED.client_id,
		    matched_at  = EXCLUDED.matched_at
		RETURNING id, created_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		mapping.IntegrationID, mapping.ExternalPhone, mapping.ClientID, mapping.MatchedAt,
	).Scan(&mapping.ID, &mapping.CreatedAt)
}

func (r *Integrations) MatchClients(ctx context.Context, integrationID int) (int, error) {
	query := `
		UPDATE integration_client_map icm
		SET client_id = bc.id, matched_at = NOW()
		FROM bot_clients bc
		WHERE bc.phone_normalized = icm.external_phone
		  AND icm.integration_id = $1
		  AND icm.client_id IS NULL`

	result, err := r.pg.DB().ExecContext(ctx, query, integrationID)
	if err != nil {
		return 0, fmt.Errorf("integrations.MatchClients: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("integrations.MatchClients rows: %w", err)
	}
	return int(rows), nil
}

func (r *Integrations) UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error {
	itemsVal, err := order.Items.Value()
	if err != nil {
		return fmt.Errorf("integrations.UpsertOrder items value: %w", err)
	}

	query := `
		INSERT INTO external_orders (integration_id, external_id, client_id, customer_phone, customer_name, items, total, ordered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (integration_id, external_id) DO UPDATE
		SET client_id       = EXCLUDED.client_id,
		    customer_phone  = EXCLUDED.customer_phone,
		    customer_name   = EXCLUDED.customer_name,
		    items           = EXCLUDED.items,
		    total           = EXCLUDED.total,
		    ordered_at      = EXCLUDED.ordered_at,
		    synced_at       = NOW()
		RETURNING id, synced_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		order.IntegrationID, order.ExternalID, order.ClientID,
		order.CustomerPhone, order.CustomerName,
		itemsVal, order.Total, order.OrderedAt,
	).Scan(&order.ID, &order.SyncedAt)
}
