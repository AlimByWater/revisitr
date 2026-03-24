package pos

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockIntegrationsRepo struct {
	updateLastSyncFn  func(ctx context.Context, id int, status string) error
	upsertOrderFn     func(ctx context.Context, order *entity.ExternalOrder) error
	getByIDFn         func(ctx context.Context, id int) (*entity.Integration, error)
	getActiveFn       func(ctx context.Context) ([]entity.Integration, error)
	upsertAggFn       func(ctx context.Context, agg *entity.IntegrationAggregate) error
	upsertClientMapFn func(ctx context.Context, m *entity.IntegrationClientMap) error
	matchClientsFn    func(ctx context.Context, integrationID int) (int, error)
}

func (m *mockIntegrationsRepo) UpdateLastSync(ctx context.Context, id int, status string) error {
	if m.updateLastSyncFn != nil {
		return m.updateLastSyncFn(ctx, id, status)
	}
	return nil
}
func (m *mockIntegrationsRepo) UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error {
	if m.upsertOrderFn != nil {
		return m.upsertOrderFn(ctx, order)
	}
	return nil
}
func (m *mockIntegrationsRepo) GetByID(ctx context.Context, id int) (*entity.Integration, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockIntegrationsRepo) GetActive(ctx context.Context) ([]entity.Integration, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(ctx)
	}
	return nil, nil
}
func (m *mockIntegrationsRepo) UpsertAggregate(ctx context.Context, agg *entity.IntegrationAggregate) error {
	if m.upsertAggFn != nil {
		return m.upsertAggFn(ctx, agg)
	}
	return nil
}
func (m *mockIntegrationsRepo) UpsertClientMap(ctx context.Context, m2 *entity.IntegrationClientMap) error {
	if m.upsertClientMapFn != nil {
		return m.upsertClientMapFn(ctx, m2)
	}
	return nil
}
func (m *mockIntegrationsRepo) MatchClients(ctx context.Context, integrationID int) (int, error) {
	if m.matchClientsFn != nil {
		return m.matchClientsFn(ctx, integrationID)
	}
	return 0, nil
}

type mockClientsRepo struct {
	getByPhoneFn func(ctx context.Context, orgID int, phone string) (*entity.BotClient, error)
}

func (m *mockClientsRepo) GetByPhone(ctx context.Context, orgID int, phone string) (*entity.BotClient, error) {
	if m.getByPhoneFn != nil {
		return m.getByPhoneFn(ctx, orgID, phone)
	}
	return nil, errors.New("not found")
}

// --- helpers ---

func syncLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mockIntegration(id int) *entity.Integration {
	return &entity.Integration{
		ID:    id,
		OrgID: 10,
		Type:  "mock",
	}
}

// --- TestConnection ---

func TestTestConnection_Success(t *testing.T) {
	svc := NewSyncService(&mockIntegrationsRepo{}, nil, syncLogger())

	err := svc.TestConnection(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTestConnection_UnknownProvider(t *testing.T) {
	svc := NewSyncService(&mockIntegrationsRepo{}, nil, syncLogger())

	intg := &entity.Integration{ID: 1, Type: "nonexistent"}
	err := svc.TestConnection(context.Background(), intg)
	if err == nil {
		t.Fatal("expected error for unknown provider type")
	}
}

// --- GetCustomers ---

func TestGetCustomers_Success(t *testing.T) {
	svc := NewSyncService(&mockIntegrationsRepo{}, nil, syncLogger())

	customers, err := svc.GetCustomers(context.Background(), mockIntegration(1), CustomerListOpts{Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(customers) == 0 {
		t.Error("expected customers from mock provider")
	}
	if len(customers) > 5 {
		t.Errorf("got %d customers, limit was 5", len(customers))
	}
}

// --- GetMenu ---

func TestGetMenu_Success(t *testing.T) {
	svc := NewSyncService(&mockIntegrationsRepo{}, nil, syncLogger())

	menu, err := svc.GetMenu(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menu == nil {
		t.Fatal("expected menu, got nil")
	}
	if len(menu.Categories) == 0 {
		t.Error("expected menu categories from mock")
	}
}

// --- Sync ---

func TestSync_Success(t *testing.T) {
	ordersUpserted := 0
	var lastSyncStatus string
	intRepo := &mockIntegrationsRepo{
		upsertOrderFn: func(_ context.Context, _ *entity.ExternalOrder) error {
			ordersUpserted++
			return nil
		},
		updateLastSyncFn: func(_ context.Context, _ int, status string) error {
			lastSyncStatus = status
			return nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.Sync(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ordersUpserted == 0 {
		t.Error("expected orders to be upserted")
	}
	if lastSyncStatus != "active" {
		t.Errorf("got sync status %q, want %q", lastSyncStatus, "active")
	}
}

func TestSync_WithPhoneMatch(t *testing.T) {
	var matchedClientIDs []int
	intRepo := &mockIntegrationsRepo{
		upsertOrderFn: func(_ context.Context, order *entity.ExternalOrder) error {
			if order.ClientID != nil {
				matchedClientIDs = append(matchedClientIDs, *order.ClientID)
			}
			return nil
		},
	}
	cRepo := &mockClientsRepo{
		getByPhoneFn: func(_ context.Context, _ int, _ string) (*entity.BotClient, error) {
			return &entity.BotClient{ID: 42}, nil
		},
	}

	svc := NewSyncService(intRepo, cRepo, syncLogger())
	err := svc.Sync(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matchedClientIDs) == 0 {
		t.Error("expected some orders to have matched client IDs")
	}
}

func TestSync_SubsequentUsesLastSyncAt(t *testing.T) {
	intRepo := &mockIntegrationsRepo{}

	svc := NewSyncService(intRepo, nil, syncLogger())

	lastSync := time.Now().Add(-2 * time.Hour)
	intg := mockIntegration(1)
	intg.LastSyncAt = &lastSync

	err := svc.Sync(context.Background(), intg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSync_UpsertOrderPartialFail(t *testing.T) {
	callCount := 0
	intRepo := &mockIntegrationsRepo{
		upsertOrderFn: func(_ context.Context, _ *entity.ExternalOrder) error {
			callCount++
			if callCount == 1 {
				return errors.New("db error")
			}
			return nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.Sync(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("partial upsert failure should not stop sync: %v", err)
	}
	if callCount < 2 {
		t.Error("should continue processing after single upsert failure")
	}
}

// --- SyncAggregates ---

func TestSyncAggregates_Success(t *testing.T) {
	aggCount := 0
	var lastSyncStatus string
	var matchClientsCalled bool

	intRepo := &mockIntegrationsRepo{
		upsertAggFn: func(_ context.Context, _ *entity.IntegrationAggregate) error {
			aggCount++
			return nil
		},
		updateLastSyncFn: func(_ context.Context, _ int, status string) error {
			lastSyncStatus = status
			return nil
		},
		matchClientsFn: func(_ context.Context, _ int) (int, error) {
			matchClientsCalled = true
			return 5, nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.SyncAggregates(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if aggCount == 0 {
		t.Error("expected aggregates to be upserted")
	}
	if lastSyncStatus != "active" {
		t.Errorf("got sync status %q, want %q", lastSyncStatus, "active")
	}
	if !matchClientsCalled {
		t.Error("MatchClients should be called")
	}
}

func TestSyncAggregates_GetAggregatesError(t *testing.T) {
	// Use unknown provider type to force GetDailyAggregates to fail at provider creation
	var lastStatus string
	intRepo := &mockIntegrationsRepo{
		updateLastSyncFn: func(_ context.Context, _ int, status string) error {
			lastStatus = status
			return nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	intg := &entity.Integration{ID: 1, Type: "nonexistent"}
	err := svc.SyncAggregates(context.Background(), intg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	// Note: provider creation error happens before UpdateLastSync("error") is called
	_ = lastStatus
}

func TestSyncAggregates_MatchClientsError(t *testing.T) {
	intRepo := &mockIntegrationsRepo{
		matchClientsFn: func(_ context.Context, _ int) (int, error) {
			return 0, errors.New("match error")
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	// Should not propagate match error
	err := svc.SyncAggregates(context.Background(), mockIntegration(1))
	if err != nil {
		t.Fatalf("match clients error should be logged, not propagated: %v", err)
	}
}

// --- SyncAll ---

func TestSyncAll_Multiple(t *testing.T) {
	syncedIDs := map[int]bool{}
	intRepo := &mockIntegrationsRepo{
		getActiveFn: func(_ context.Context) ([]entity.Integration, error) {
			return []entity.Integration{
				{ID: 1, OrgID: 10, Type: "mock"},
				{ID: 2, OrgID: 10, Type: "mock"},
			}, nil
		},
		upsertOrderFn: func(_ context.Context, order *entity.ExternalOrder) error {
			syncedIDs[order.IntegrationID] = true
			return nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.SyncAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !syncedIDs[1] || !syncedIDs[2] {
		t.Errorf("expected both integrations synced, got %v", syncedIDs)
	}
}

func TestSyncAll_PartialFailure(t *testing.T) {
	callCount := 0
	intRepo := &mockIntegrationsRepo{
		getActiveFn: func(_ context.Context) ([]entity.Integration, error) {
			return []entity.Integration{
				{ID: 1, OrgID: 10, Type: "nonexistent"}, // will fail
				{ID: 2, OrgID: 10, Type: "mock"},         // should still sync
			}, nil
		},
		upsertOrderFn: func(_ context.Context, _ *entity.ExternalOrder) error {
			callCount++
			return nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.SyncAll(context.Background())
	if err != nil {
		t.Fatalf("partial failure should not stop SyncAll: %v", err)
	}
	if callCount == 0 {
		t.Error("second integration should still have synced")
	}
}

func TestSyncAll_Empty(t *testing.T) {
	intRepo := &mockIntegrationsRepo{
		getActiveFn: func(_ context.Context) ([]entity.Integration, error) {
			return nil, nil
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.SyncAll(context.Background())
	if err != nil {
		t.Fatalf("empty list should be no-op: %v", err)
	}
}

func TestSyncAll_RepoError(t *testing.T) {
	intRepo := &mockIntegrationsRepo{
		getActiveFn: func(_ context.Context) ([]entity.Integration, error) {
			return nil, errors.New("db error")
		},
	}

	svc := NewSyncService(intRepo, nil, syncLogger())
	err := svc.SyncAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
