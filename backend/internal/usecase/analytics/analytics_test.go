package analytics

import (
	"context"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockAnalyticsRepo struct {
	getSalesMetricsFn  func(ctx context.Context, f entity.AnalyticsFilter) (*entity.SalesMetrics, error)
	getSalesChartsFn   func(ctx context.Context, f entity.AnalyticsFilter) (map[string][]entity.SalesChartPoint, error)
	getLoyaltyFn       func(ctx context.Context, f entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error)
	getCampaignFn      func(ctx context.Context, f entity.AnalyticsFilter) (*entity.CampaignAnalytics, error)
	refreshMVFn        func(ctx context.Context) error
}

func (m *mockAnalyticsRepo) GetSalesMetrics(ctx context.Context, f entity.AnalyticsFilter) (*entity.SalesMetrics, error) {
	return m.getSalesMetricsFn(ctx, f)
}
func (m *mockAnalyticsRepo) GetSalesCharts(ctx context.Context, f entity.AnalyticsFilter) (map[string][]entity.SalesChartPoint, error) {
	return m.getSalesChartsFn(ctx, f)
}
func (m *mockAnalyticsRepo) GetLoyaltyAnalytics(ctx context.Context, f entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error) {
	return m.getLoyaltyFn(ctx, f)
}
func (m *mockAnalyticsRepo) GetCampaignAnalytics(ctx context.Context, f entity.AnalyticsFilter) (*entity.CampaignAnalytics, error) {
	return m.getCampaignFn(ctx, f)
}
func (m *mockAnalyticsRepo) RefreshMaterializedViews(ctx context.Context) error {
	return m.refreshMVFn(ctx)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestGetSalesAnalytics_DefaultDates(t *testing.T) {
	called := false
	repo := &mockAnalyticsRepo{
		getSalesMetricsFn: func(_ context.Context, f entity.AnalyticsFilter) (*entity.SalesMetrics, error) {
			called = true
			if f.From.IsZero() {
				t.Error("expected From to be set by default")
			}
			if f.To.IsZero() {
				t.Error("expected To to be set by default")
			}
			return &entity.SalesMetrics{}, nil
		},
		getSalesChartsFn: func(_ context.Context, _ entity.AnalyticsFilter) (map[string][]entity.SalesChartPoint, error) {
			return map[string][]entity.SalesChartPoint{}, nil
		},
	}

	uc := New(repo)
	_, err := uc.GetSalesAnalytics(context.Background(), 1, entity.AnalyticsFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("repo was not called")
	}
}

func TestGetSalesAnalytics_InvalidDateRange(t *testing.T) {
	repo := &mockAnalyticsRepo{}
	uc := New(repo)

	from := time.Now()
	to := time.Now().AddDate(0, 0, -1) // to before from

	_, err := uc.GetSalesAnalytics(context.Background(), 1, entity.AnalyticsFilter{
		From: from,
		To:   to,
	})

	if err != ErrInvalidDateRange {
		t.Fatalf("expected ErrInvalidDateRange, got: %v", err)
	}
}

func TestGetSalesAnalytics_OrgInjected(t *testing.T) {
	const wantOrgID = 42
	repo := &mockAnalyticsRepo{
		getSalesMetricsFn: func(_ context.Context, f entity.AnalyticsFilter) (*entity.SalesMetrics, error) {
			if f.OrgID != wantOrgID {
				t.Errorf("expected org_id=%d, got %d", wantOrgID, f.OrgID)
			}
			return &entity.SalesMetrics{}, nil
		},
		getSalesChartsFn: func(_ context.Context, _ entity.AnalyticsFilter) (map[string][]entity.SalesChartPoint, error) {
			return map[string][]entity.SalesChartPoint{}, nil
		},
	}

	uc := New(repo)
	_, err := uc.GetSalesAnalytics(context.Background(), wantOrgID, entity.AnalyticsFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
