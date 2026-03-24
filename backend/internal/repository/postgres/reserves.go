package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Reserves struct {
	pg *Module
}

func NewReserves(pg *Module) *Reserves {
	return &Reserves{pg: pg}
}

func (r *Reserves) CreateReserve(ctx context.Context, reserve *entity.BalanceReserve) error {
	query := `
		INSERT INTO balance_reserves (client_id, program_id, amount, status, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.pg.DB().QueryRowContext(ctx, query,
		reserve.ClientID, reserve.ProgramID, reserve.Amount, reserve.Status, reserve.ExpiresAt,
	).Scan(&reserve.ID, &reserve.CreatedAt)
	if err != nil {
		return fmt.Errorf("reserves.CreateReserve: %w", err)
	}
	return nil
}

func (r *Reserves) GetReserve(ctx context.Context, id int) (*entity.BalanceReserve, error) {
	var reserve entity.BalanceReserve
	err := r.pg.DB().GetContext(ctx, &reserve, "SELECT * FROM balance_reserves WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reserves.GetReserve: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("reserves.GetReserve: %w", err)
	}
	return &reserve, nil
}

func (r *Reserves) UpdateReserve(ctx context.Context, reserve *entity.BalanceReserve) error {
	query := `UPDATE balance_reserves SET status = $1 WHERE id = $2`
	result, err := r.pg.DB().ExecContext(ctx, query, reserve.Status, reserve.ID)
	if err != nil {
		return fmt.Errorf("reserves.UpdateReserve: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reserves.UpdateReserve rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reserves.UpdateReserve: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Reserves) GetPendingReserves(ctx context.Context, clientID, programID int) ([]entity.BalanceReserve, error) {
	var reserves []entity.BalanceReserve
	query := `
		SELECT * FROM balance_reserves
		WHERE client_id = $1 AND program_id = $2 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at`
	err := r.pg.DB().SelectContext(ctx, &reserves, query, clientID, programID)
	if err != nil {
		return nil, fmt.Errorf("reserves.GetPendingReserves: %w", err)
	}
	return reserves, nil
}

func (r *Reserves) ExpireOldReserves(ctx context.Context) (int, error) {
	query := `UPDATE balance_reserves SET status = 'expired' WHERE status = 'pending' AND expires_at <= NOW()`
	result, err := r.pg.DB().ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("reserves.ExpireOldReserves: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("reserves.ExpireOldReserves rows: %w", err)
	}
	return int(rows), nil
}
