package dashboard_test

import (
	"context"
	"errors"
	"testing"

	"revisitr/internal/entity"
	"revisitr/internal/usecase/dashboard"
)

// --- mock ---

type mockRepo struct {
	getWidgetsFn func(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error)
	getChartsFn  func(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error)
}

func (m *mockRepo) GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error) {
	return m.getWidgetsFn(ctx, orgID, filter)
}
func (m *mockRepo) GetCharts(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error) {
	return m.getChartsFn(ctx, orgID, filter)
}

// --- tests ---

func TestGetWidgets_ReturnsWidgets(t *testing.T) {
	expected := &entity.DashboardWidgets{}
	expected.NewClients.Value = 10
	repo := &mockRepo{
		getWidgetsFn: func(_ context.Context, _ int, _ entity.DashboardFilter) (*entity.DashboardWidgets, error) {
			return expected, nil
		},
	}
	uc := dashboard.New(repo)

	got, err := uc.GetWidgets(context.Background(), 1, entity.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.NewClients.Value != 10 {
		t.Errorf("expected NewClients.Value=10, got %v", got.NewClients.Value)
	}
}

func TestGetWidgets_PropagatesError(t *testing.T) {
	repo := &mockRepo{
		getWidgetsFn: func(_ context.Context, _ int, _ entity.DashboardFilter) (*entity.DashboardWidgets, error) {
			return nil, errors.New("db error")
		},
	}
	uc := dashboard.New(repo)

	_, err := uc.GetWidgets(context.Background(), 1, entity.DashboardFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetCharts_ReturnsCharts(t *testing.T) {
	expected := &entity.DashboardCharts{}
	repo := &mockRepo{
		getChartsFn: func(_ context.Context, _ int, _ entity.DashboardFilter) (*entity.DashboardCharts, error) {
			return expected, nil
		},
	}
	uc := dashboard.New(repo)

	got, err := uc.GetCharts(context.Background(), 1, entity.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Error("expected non-nil charts")
	}
}

func TestGetCharts_PropagatesError(t *testing.T) {
	repo := &mockRepo{
		getChartsFn: func(_ context.Context, _ int, _ entity.DashboardFilter) (*entity.DashboardCharts, error) {
			return nil, errors.New("db error")
		},
	}
	uc := dashboard.New(repo)

	_, err := uc.GetCharts(context.Background(), 1, entity.DashboardFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
