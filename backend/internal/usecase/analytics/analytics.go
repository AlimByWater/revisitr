package analytics

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type analyticsRepo interface {
	GetSalesMetrics(ctx context.Context, f entity.AnalyticsFilter) (*entity.SalesMetrics, error)
	GetSalesCharts(ctx context.Context, f entity.AnalyticsFilter) (map[string][]entity.SalesChartPoint, error)
	GetLoyaltyAnalytics(ctx context.Context, f entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error)
	GetCampaignAnalytics(ctx context.Context, f entity.AnalyticsFilter) (*entity.CampaignAnalytics, error)
	RefreshMaterializedViews(ctx context.Context) error
}

type Usecase struct {
	logger *slog.Logger
	repo   analyticsRepo
}

func New(repo analyticsRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) GetSalesAnalytics(ctx context.Context, orgID int, f entity.AnalyticsFilter) (*entity.SalesAnalytics, error) {
	f.OrgID = orgID
	if err := validateFilter(&f); err != nil {
		return nil, err
	}

	metrics, err := uc.repo.GetSalesMetrics(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("get sales metrics: %w", err)
	}

	charts, err := uc.repo.GetSalesCharts(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("get sales charts: %w", err)
	}

	return &entity.SalesAnalytics{
		Metrics: *metrics,
		Charts:  charts,
	}, nil
}

func (uc *Usecase) GetLoyaltyAnalytics(ctx context.Context, orgID int, f entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error) {
	f.OrgID = orgID
	if err := validateFilter(&f); err != nil {
		return nil, err
	}

	result, err := uc.repo.GetLoyaltyAnalytics(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("get loyalty analytics: %w", err)
	}

	return result, nil
}

func (uc *Usecase) GetCampaignAnalytics(ctx context.Context, orgID int, f entity.AnalyticsFilter) (*entity.CampaignAnalytics, error) {
	f.OrgID = orgID
	if err := validateFilter(&f); err != nil {
		return nil, err
	}

	result, err := uc.repo.GetCampaignAnalytics(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("get campaign analytics: %w", err)
	}

	return result, nil
}

// validateFilter fills zero-value dates with defaults: last 30 days.
func validateFilter(f *entity.AnalyticsFilter) error {
	if f.From.IsZero() {
		f.From = time.Now().AddDate(0, 0, -30)
	}
	if f.To.IsZero() {
		f.To = time.Now()
	}
	if f.To.Before(f.From) {
		return ErrInvalidDateRange
	}
	return nil
}
