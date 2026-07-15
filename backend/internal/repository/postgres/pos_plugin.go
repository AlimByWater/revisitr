package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type PluginKeys struct {
	pg *Module
}

func NewPluginKeys(pg *Module) *PluginKeys {
	return &PluginKeys{pg: pg}
}

func (r *PluginKeys) Create(ctx context.Context, k *entity.PluginKey) error {
	query := `
		INSERT INTO pos_plugin_keys (org_id, integration_id, key_hash, label)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		k.OrgID, k.IntegrationID, k.KeyHash, k.Label,
	).Scan(&k.ID, &k.CreatedAt)
}

func (r *PluginKeys) GetActiveByHash(ctx context.Context, keyHash string) (*entity.PluginKey, error) {
	var k entity.PluginKey
	err := r.pg.DB().GetContext(ctx, &k,
		"SELECT * FROM pos_plugin_keys WHERE key_hash = $1 AND revoked_at IS NULL", keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pos_plugin_keys.GetActiveByHash: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("pos_plugin_keys.GetActiveByHash: %w", err)
	}
	return &k, nil
}

func (r *PluginKeys) ListByIntegration(ctx context.Context, integrationID int) ([]entity.PluginKey, error) {
	var keys []entity.PluginKey
	err := r.pg.DB().SelectContext(ctx, &keys,
		"SELECT * FROM pos_plugin_keys WHERE integration_id = $1 ORDER BY created_at DESC", integrationID)
	if err != nil {
		return nil, fmt.Errorf("pos_plugin_keys.ListByIntegration: %w", err)
	}
	return keys, nil
}

func (r *PluginKeys) TouchLastUsed(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE pos_plugin_keys SET last_used_at = NOW() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("pos_plugin_keys.TouchLastUsed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("pos_plugin_keys.TouchLastUsed rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pos_plugin_keys.TouchLastUsed: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *PluginKeys) Revoke(ctx context.Context, id, orgID int) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE pos_plugin_keys SET revoked_at = NOW() WHERE id = $1 AND org_id = $2", id, orgID)
	if err != nil {
		return fmt.Errorf("pos_plugin_keys.Revoke: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("pos_plugin_keys.Revoke rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pos_plugin_keys.Revoke: %w", sql.ErrNoRows)
	}
	return nil
}

type PluginOperations struct {
	pg *Module
}

func NewPluginOperations(pg *Module) *PluginOperations {
	return &PluginOperations{pg: pg}
}

func (r *PluginOperations) Get(ctx context.Context, integrationID int, externalOrderID, opType string) (*entity.PluginOperation, error) {
	var op entity.PluginOperation
	err := r.pg.DB().GetContext(ctx, &op,
		"SELECT * FROM pos_plugin_operations WHERE integration_id = $1 AND external_order_id = $2 AND op_type = $3",
		integrationID, externalOrderID, opType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pos_plugin_operations.Get: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("pos_plugin_operations.Get: %w", err)
	}
	return &op, nil
}

func (r *PluginOperations) Insert(ctx context.Context, op *entity.PluginOperation) error {
	query := `
		INSERT INTO pos_plugin_operations (integration_id, external_order_id, op_type, client_id, program_id, amount, balance_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return r.pg.DB().QueryRowContext(ctx, query,
		op.IntegrationID, op.ExternalOrderID, op.OpType, op.ClientID, op.ProgramID, op.Amount, op.BalanceAfter,
	).Scan(&op.ID, &op.CreatedAt)
}
