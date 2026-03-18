package dashboard

import (
	"context"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type dashboardRepo interface {
	GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error)
	GetCharts(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error)
}

type Usecase struct {
	logger *slog.Logger
	repo   dashboardRepo
}

func New(repo dashboardRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error) {
	widgets, err := uc.repo.GetWidgets(ctx, orgID, filter)
	if err != nil {
		return nil, fmt.Errorf("get dashboard widgets: %w", err)
	}
	return widgets, nil
}

func (uc *Usecase) GetCharts(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error) {
	charts, err := uc.repo.GetCharts(ctx, orgID, filter)
	if err != nil {
		return nil, fmt.Errorf("get dashboard charts: %w", err)
	}
	return charts, nil
}
