package clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type clientsRepo interface {
	GetByOrgID(ctx context.Context, orgID int, filter entity.ClientFilter) ([]entity.ClientProfile, int, error)
	GetByID(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error)
	Update(ctx context.Context, orgID, clientID int, req *entity.UpdateClientRequest) error
	GetStats(ctx context.Context, orgID int) (*entity.ClientStats, error)
	CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error)
	GetTransactionsByClientID(ctx context.Context, clientID int, limit, offset int) ([]entity.LoyaltyTransaction, error)
}

type Usecase struct {
	logger *slog.Logger
	repo   clientsRepo
}

func New(repo clientsRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) List(ctx context.Context, orgID int, filter entity.ClientFilter) (*entity.PaginatedResponse[entity.ClientProfile], error) {
	profiles, total, err := uc.repo.GetByOrgID(ctx, orgID, filter)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	return &entity.PaginatedResponse[entity.ClientProfile]{
		Items: profiles,
		Total: total,
	}, nil
}

func (uc *Usecase) GetProfile(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error) {
	profile, err := uc.repo.GetByID(ctx, orgID, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, fmt.Errorf("get client profile: %w", err)
	}

	txs, err := uc.repo.GetTransactionsByClientID(ctx, clientID, 50, 0)
	if err != nil {
		return nil, fmt.Errorf("get client transactions: %w", err)
	}
	profile.Transactions = txs

	return profile, nil
}

func (uc *Usecase) UpdateTags(ctx context.Context, orgID, clientID int, req *entity.UpdateClientRequest) error {
	if err := uc.repo.Update(ctx, orgID, clientID, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrClientNotFound
		}
		return fmt.Errorf("update client tags: %w", err)
	}

	return nil
}

func (uc *Usecase) GetStats(ctx context.Context, orgID int) (*entity.ClientStats, error) {
	stats, err := uc.repo.GetStats(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get client stats: %w", err)
	}

	return stats, nil
}

func (uc *Usecase) CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error) {
	count, err := uc.repo.CountByFilter(ctx, orgID, filter)
	if err != nil {
		return 0, fmt.Errorf("count clients by filter: %w", err)
	}

	return count, nil
}
