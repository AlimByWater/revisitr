package orders

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockOrdersRepo struct {
	listByBotFn    func(ctx context.Context, botID int, source, status string) ([]entity.Order, error)
	listByOrgFn    func(ctx context.Context, orgID int, source, status string) ([]entity.Order, error)
	updateStatusFn func(ctx context.Context, orderID int, status string) error
	getOrgIDFn     func(ctx context.Context, orderID int) (int, error)
}

func (m *mockOrdersRepo) ListByBot(ctx context.Context, botID int, source, status string) ([]entity.Order, error) {
	if m.listByBotFn != nil {
		return m.listByBotFn(ctx, botID, source, status)
	}
	return nil, nil
}
func (m *mockOrdersRepo) ListByOrg(ctx context.Context, orgID int, source, status string) ([]entity.Order, error) {
	if m.listByOrgFn != nil {
		return m.listByOrgFn(ctx, orgID, source, status)
	}
	return nil, nil
}
func (m *mockOrdersRepo) UpdateStatus(ctx context.Context, orderID int, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, orderID, status)
	}
	return nil
}
func (m *mockOrdersRepo) GetOrgID(ctx context.Context, orderID int) (int, error) {
	if m.getOrgIDFn != nil {
		return m.getOrgIDFn(ctx, orderID)
	}
	return 0, nil
}

type mockBotsGetter struct {
	getByIDFn func(ctx context.Context, id int) (*entity.Bot, error)
}

func (m *mockBotsGetter) GetByID(ctx context.Context, id int) (*entity.Bot, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &entity.Bot{ID: id, OrgID: 1}, nil
}

func newTestUsecase(t *testing.T, repo *mockOrdersRepo, bots *mockBotsGetter) *Usecase {
	t.Helper()
	if repo == nil {
		repo = &mockOrdersRepo{}
	}
	if bots == nil {
		bots = &mockBotsGetter{}
	}
	uc := New(repo, bots)
	if err := uc.Init(context.Background(), slog.Default()); err != nil {
		t.Fatalf("init usecase: %v", err)
	}
	return uc
}

// --- tests ---

func TestUpdateOrderStatusInvalid(t *testing.T) {
	uc := newTestUsecase(t, &mockOrdersRepo{}, nil)
	err := uc.UpdateOrderStatus(context.Background(), 1, 3, "cooking")
	if !errors.Is(err, ErrValidation) {
		t.Errorf("unknown status: expected ErrValidation, got %v", err)
	}
}

func TestUpdateOrderStatusForeignOrg(t *testing.T) {
	repo := &mockOrdersRepo{
		getOrgIDFn: func(_ context.Context, _ int) (int, error) { return 99, nil },
	}
	uc := newTestUsecase(t, repo, nil)
	err := uc.UpdateOrderStatus(context.Background(), 1, 3, entity.OrderStatusCancelled)
	if !errors.Is(err, ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestUpdateOrderStatusMissing(t *testing.T) {
	uc := newTestUsecase(t, &mockOrdersRepo{}, nil)
	err := uc.UpdateOrderStatus(context.Background(), 1, 3, entity.OrderStatusSent)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListOrdersForeignBot(t *testing.T) {
	bots := &mockBotsGetter{
		getByIDFn: func(_ context.Context, id int) (*entity.Bot, error) {
			return &entity.Bot{ID: id, OrgID: 99}, nil
		},
	}
	uc := newTestUsecase(t, nil, bots)
	if _, err := uc.ListOrders(context.Background(), 1, 10, "", ""); !errors.Is(err, ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestListOrgOrdersPassesOrg(t *testing.T) {
	var gotOrgID int
	repo := &mockOrdersRepo{
		listByOrgFn: func(_ context.Context, orgID int, _, _ string) ([]entity.Order, error) {
			gotOrgID = orgID
			return []entity.Order{{ID: 1}}, nil
		},
	}
	uc := newTestUsecase(t, repo, nil)
	orders, err := uc.ListOrgOrders(context.Background(), 7, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 1 || gotOrgID != 7 {
		t.Errorf("expected 1 order for org 7, got %d orders, org %d", len(orders), gotOrgID)
	}
}

func TestListOrdersPassesFilters(t *testing.T) {
	var gotSource, gotStatus string
	repo := &mockOrdersRepo{
		listByBotFn: func(_ context.Context, _ int, source, status string) ([]entity.Order, error) {
			gotSource, gotStatus = source, status
			return []entity.Order{{ID: 1}}, nil
		},
	}
	uc := newTestUsecase(t, repo, nil)
	orders, err := uc.ListOrders(context.Background(), 1, 10, entity.OrderSourceLunch, entity.OrderStatusNew)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}
	if gotSource != entity.OrderSourceLunch || gotStatus != entity.OrderStatusNew {
		t.Errorf("filters not passed through: source=%q status=%q", gotSource, gotStatus)
	}
}
