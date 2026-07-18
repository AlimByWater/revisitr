package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrNotFound   = errors.New("order not found")
	ErrNotOwner   = errors.New("not the owner")
	ErrValidation = errors.New("validation failed")
)

type ordersRepo interface {
	ListByBot(ctx context.Context, botID int, source, status string) ([]entity.Order, error)
	ListByOrg(ctx context.Context, orgID int, source, status string) ([]entity.Order, error)
	UpdateStatus(ctx context.Context, orderID int, status string) error
	GetOrgID(ctx context.Context, orderID int) (int, error)
}

type botsGetter interface {
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type Usecase struct {
	logger *slog.Logger
	repo   ordersRepo
	bots   botsGetter
}

func New(repo ordersRepo, bots botsGetter) *Usecase {
	return &Usecase{repo: repo, bots: bots}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// checkBot verifies the bot exists and belongs to the org.
func (uc *Usecase) checkBot(ctx context.Context, orgID, botID int) error {
	bot, err := uc.bots.GetByID(ctx, botID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if bot.OrgID != orgID {
		return ErrNotOwner
	}
	return nil
}

func (uc *Usecase) ListOrders(ctx context.Context, orgID, botID int, source, status string) ([]entity.Order, error) {
	if err := uc.checkBot(ctx, orgID, botID); err != nil {
		return nil, err
	}
	return uc.repo.ListByBot(ctx, botID, source, status)
}

func (uc *Usecase) ListOrgOrders(ctx context.Context, orgID int, source, status string) ([]entity.Order, error) {
	return uc.repo.ListByOrg(ctx, orgID, source, status)
}

func (uc *Usecase) UpdateOrderStatus(ctx context.Context, orgID, orderID int, status string) error {
	switch status {
	case entity.OrderStatusNew, entity.OrderStatusSent, entity.OrderStatusCancelled:
	default:
		return fmt.Errorf("%w: unknown status %q", ErrValidation, status)
	}
	orderOrg, err := uc.repo.GetOrgID(ctx, orderID)
	if err != nil {
		return err
	}
	if orderOrg == 0 {
		return ErrNotFound
	}
	if orderOrg != orgID {
		return ErrNotOwner
	}
	return uc.repo.UpdateStatus(ctx, orderID, status)
}
