package loyalty

import (
	"context"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

// ReservePoints reserves bonus points for later spending at POS.
// Returns reserve ID for confirmation or cancellation.
func (uc *Usecase) ReservePoints(ctx context.Context, clientID, programID int, amount float64) (int, error) {
	cl, err := uc.getOrCreateClientLoyalty(ctx, clientID, programID)
	if err != nil {
		return 0, err
	}

	// Calculate available balance (minus pending reserves)
	pending, err := uc.repo.GetPendingReserves(ctx, clientID, programID)
	if err != nil {
		return 0, fmt.Errorf("ReservePoints get pending: %w", err)
	}
	var reserved float64
	for _, r := range pending {
		reserved += r.Amount
	}
	available := cl.Balance - reserved

	if available < amount {
		return 0, ErrInsufficientPoints
	}

	reserve := &entity.BalanceReserve{
		ClientID:  clientID,
		ProgramID: programID,
		Amount:    amount,
		Status:    "pending",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := uc.repo.CreateReserve(ctx, reserve); err != nil {
		return 0, fmt.Errorf("ReservePoints: %w", err)
	}

	return reserve.ID, nil
}

// ConfirmReserve confirms a pending reserve and spends the points.
func (uc *Usecase) ConfirmReserve(ctx context.Context, reserveID int) (*entity.ClientLoyalty, error) {
	reserve, err := uc.repo.GetReserve(ctx, reserveID)
	if err != nil {
		return nil, fmt.Errorf("ConfirmReserve: %w", err)
	}

	if reserve.Status != "pending" {
		return nil, ErrReserveNotPending
	}
	if time.Now().After(reserve.ExpiresAt) {
		return nil, ErrReserveExpired
	}

	cl, err := uc.SpendPoints(ctx, reserve.ClientID, reserve.ProgramID, reserve.Amount, "Списание по резерву")
	if err != nil {
		return nil, err
	}

	reserve.Status = "confirmed"
	if err := uc.repo.UpdateReserve(ctx, reserve); err != nil {
		return nil, fmt.Errorf("ConfirmReserve update: %w", err)
	}

	return cl, nil
}

// CancelReserve cancels a pending reserve.
func (uc *Usecase) CancelReserve(ctx context.Context, reserveID int) error {
	reserve, err := uc.repo.GetReserve(ctx, reserveID)
	if err != nil {
		return fmt.Errorf("CancelReserve: %w", err)
	}

	reserve.Status = "cancelled"
	return uc.repo.UpdateReserve(ctx, reserve)
}
