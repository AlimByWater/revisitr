package integrations

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockIntegrationsRepo struct {
	createFn        func(ctx context.Context, intg *entity.Integration) error
	getByIDFn       func(ctx context.Context, id int) (*entity.Integration, error)
	getByOrgIDFn    func(ctx context.Context, orgID int) ([]entity.Integration, error)
	updateFn        func(ctx context.Context, intg *entity.Integration) error
	deleteFn        func(ctx context.Context, id int) error
	updateLastSyncFn func(ctx context.Context, id int, status string) error
	upsertOrderFn   func(ctx context.Context, order *entity.ExternalOrder) error
}

func (m *mockIntegrationsRepo) Create(ctx context.Context, intg *entity.Integration) error {
	return m.createFn(ctx, intg)
}
func (m *mockIntegrationsRepo) GetByID(ctx context.Context, id int) (*entity.Integration, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockIntegrationsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Integration, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockIntegrationsRepo) Update(ctx context.Context, intg *entity.Integration) error {
	return m.updateFn(ctx, intg)
}
func (m *mockIntegrationsRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockIntegrationsRepo) UpdateLastSync(ctx context.Context, id int, status string) error {
	return m.updateLastSyncFn(ctx, id, status)
}
func (m *mockIntegrationsRepo) UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error {
	return m.upsertOrderFn(ctx, order)
}

type mockSyncService struct {
	syncFn func(ctx context.Context, integration *entity.Integration) error
}

func (m *mockSyncService) Sync(ctx context.Context, integration *entity.Integration) error {
	return m.syncFn(ctx, integration)
}

func testIntegration(id, orgID int) *entity.Integration {
	return &entity.Integration{
		ID:        id,
		OrgID:     orgID,
		Type:      "iiko",
		Status:    "inactive",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreate_SetsInactiveStatus(t *testing.T) {
	repo := &mockIntegrationsRepo{
		createFn: func(_ context.Context, intg *entity.Integration) error {
			intg.ID = 1
			return nil
		},
	}
	uc := New(repo, &mockSyncService{})

	intg, err := uc.Create(context.Background(), 10, &entity.CreateIntegrationRequest{
		Type: "iiko",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intg.Status != "inactive" {
		t.Errorf("expected status=inactive, got %s", intg.Status)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockIntegrationsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Integration, error) {
			return nil, fmt.Errorf("integrations.GetByID: %w", sql.ErrNoRows)
		},
	}
	uc := New(repo, &mockSyncService{})

	_, err := uc.GetByID(context.Background(), 99, 1)
	if err != ErrIntegrationNotFound {
		t.Errorf("expected ErrIntegrationNotFound, got: %v", err)
	}
}

func TestGetByID_WrongOrg(t *testing.T) {
	repo := &mockIntegrationsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Integration, error) {
			return testIntegration(id, 5), nil
		},
	}
	uc := New(repo, &mockSyncService{})

	_, err := uc.GetByID(context.Background(), 1, 99)
	if err != ErrNotIntegrationOwner {
		t.Errorf("expected ErrNotIntegrationOwner, got: %v", err)
	}
}

func TestSyncNow_CallsSyncService(t *testing.T) {
	syncCalled := false
	repo := &mockIntegrationsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Integration, error) {
			return testIntegration(id, 1), nil
		},
	}
	syncSvc := &mockSyncService{
		syncFn: func(_ context.Context, _ *entity.Integration) error {
			syncCalled = true
			return nil
		},
	}
	uc := New(repo, syncSvc)

	if err := uc.SyncNow(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !syncCalled {
		t.Error("expected sync service to be called")
	}
}
