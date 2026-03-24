package pos

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockPOSRepo struct {
	createFn     func(ctx context.Context, pos *entity.POSLocation) error
	getByIDFn    func(ctx context.Context, id int) (*entity.POSLocation, error)
	getByOrgIDFn func(ctx context.Context, orgID int) ([]entity.POSLocation, error)
	updateFn     func(ctx context.Context, pos *entity.POSLocation) error
	deleteFn     func(ctx context.Context, id int) error
}

func (m *mockPOSRepo) Create(ctx context.Context, pos *entity.POSLocation) error {
	return m.createFn(ctx, pos)
}
func (m *mockPOSRepo) GetByID(ctx context.Context, id int) (*entity.POSLocation, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockPOSRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.POSLocation, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockPOSRepo) Update(ctx context.Context, pos *entity.POSLocation) error {
	return m.updateFn(ctx, pos)
}
func (m *mockPOSRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}

// --- helpers ---

func newTestUC(repo posRepo) *Usecase {
	uc := New(repo)
	_ = uc.Init(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	return uc
}

func testPOS(id, orgID int) *entity.POSLocation {
	return &entity.POSLocation{
		ID:       id,
		OrgID:    orgID,
		Name:     "Test Location",
		Address:  "123 Main St",
		Phone:    "+71234567890",
		Schedule: entity.Schedule{"monday": {Open: "09:00", Close: "22:00"}},
		IsActive: true,
	}
}

func ptr[T any](v T) *T { return &v }

// --- Create ---

func TestCreate_Success(t *testing.T) {
	repo := &mockPOSRepo{
		createFn: func(_ context.Context, p *entity.POSLocation) error {
			p.ID = 42
			return nil
		},
	}
	uc := newTestUC(repo)

	pos, err := uc.Create(context.Background(), 10, &entity.CreatePOSRequest{
		Name:    "My Location",
		Address: "456 Oak Ave",
		Phone:   "+79991234567",
		Schedule: entity.Schedule{
			"monday": {Open: "10:00", Close: "21:00"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos.ID != 42 {
		t.Errorf("got ID %d, want 42", pos.ID)
	}
	if pos.OrgID != 10 {
		t.Errorf("got OrgID %d, want 10", pos.OrgID)
	}
	if !pos.IsActive {
		t.Error("new POS should be active by default")
	}
}

func TestCreate_NilSchedule(t *testing.T) {
	repo := &mockPOSRepo{
		createFn: func(_ context.Context, p *entity.POSLocation) error {
			p.ID = 1
			return nil
		},
	}
	uc := newTestUC(repo)

	pos, err := uc.Create(context.Background(), 10, &entity.CreatePOSRequest{
		Name: "No Schedule",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos.Schedule == nil {
		t.Error("nil schedule should be defaulted to empty Schedule")
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := &mockPOSRepo{
		createFn: func(_ context.Context, _ *entity.POSLocation) error {
			return errors.New("db error")
		},
	}
	uc := newTestUC(repo)

	_, err := uc.Create(context.Background(), 10, &entity.CreatePOSRequest{Name: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetByOrgID ---

func TestGetByOrgID_Success(t *testing.T) {
	repo := &mockPOSRepo{
		getByOrgIDFn: func(_ context.Context, orgID int) ([]entity.POSLocation, error) {
			return []entity.POSLocation{
				{ID: 1, OrgID: orgID},
				{ID: 2, OrgID: orgID},
			}, nil
		},
	}
	uc := newTestUC(repo)

	locs, err := uc.GetByOrgID(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 2 {
		t.Errorf("got %d locations, want 2", len(locs))
	}
}

func TestGetByOrgID_RepoError(t *testing.T) {
	repo := &mockPOSRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.POSLocation, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newTestUC(repo)

	_, err := uc.GetByOrgID(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetByID ---

func TestGetByID_Success(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
	}
	uc := newTestUC(repo)

	got, err := uc.GetByID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("got ID %d, want 1", got.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return nil, sql.ErrNoRows
		},
	}
	uc := newTestUC(repo)

	_, err := uc.GetByID(context.Background(), 999, 10)
	if !errors.Is(err, ErrPOSNotFound) {
		t.Errorf("got %v, want ErrPOSNotFound", err)
	}
}

func TestGetByID_WrongOrg(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
	}
	uc := newTestUC(repo)

	_, err := uc.GetByID(context.Background(), 1, 999)
	if !errors.Is(err, ErrNotPOSOwner) {
		t.Errorf("got %v, want ErrNotPOSOwner", err)
	}
}

func TestGetByID_OtherRepoError(t *testing.T) {
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return nil, errors.New("connection error")
		},
	}
	uc := newTestUC(repo)

	_, err := uc.GetByID(context.Background(), 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, ErrPOSNotFound) {
		t.Error("non-ErrNoRows should not map to ErrPOSNotFound")
	}
}

// --- Update ---

func TestUpdate_NameOnly(t *testing.T) {
	pos := testPOS(1, 10)
	fetchCount := 0
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			fetchCount++
			return pos, nil
		},
		updateFn: func(_ context.Context, _ *entity.POSLocation) error {
			return nil
		},
	}
	uc := newTestUC(repo)

	got, err := uc.Update(context.Background(), 1, 10, &entity.UpdatePOSRequest{
		Name: ptr("Updated"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("got Name %q, want %q", got.Name, "Updated")
	}
	// Should fetch twice: once for auth check, once for refetch
	if fetchCount != 2 {
		t.Errorf("GetByID called %d times, want 2 (auth + refetch)", fetchCount)
	}
}

func TestUpdate_IsActive(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
		updateFn: func(_ context.Context, p *entity.POSLocation) error {
			if p.IsActive {
				t.Error("IsActive should be false after update")
			}
			return nil
		},
	}
	uc := newTestUC(repo)

	_, err := uc.Update(context.Background(), 1, 10, &entity.UpdatePOSRequest{
		IsActive: ptr(false),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return nil, sql.ErrNoRows
		},
	}
	uc := newTestUC(repo)

	_, err := uc.Update(context.Background(), 999, 10, &entity.UpdatePOSRequest{Name: ptr("x")})
	if !errors.Is(err, ErrPOSNotFound) {
		t.Errorf("got %v, want ErrPOSNotFound", err)
	}
}

func TestUpdate_WrongOrg(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
	}
	uc := newTestUC(repo)

	_, err := uc.Update(context.Background(), 1, 999, &entity.UpdatePOSRequest{Name: ptr("x")})
	if !errors.Is(err, ErrNotPOSOwner) {
		t.Errorf("got %v, want ErrNotPOSOwner", err)
	}
}

func TestUpdate_RepoError(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
		updateFn: func(_ context.Context, _ *entity.POSLocation) error {
			return errors.New("db error")
		},
	}
	uc := newTestUC(repo)

	_, err := uc.Update(context.Background(), 1, 10, &entity.UpdatePOSRequest{Name: ptr("x")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	pos := testPOS(1, 10)
	deleted := false
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
		deleteFn: func(_ context.Context, _ int) error {
			deleted = true
			return nil
		},
	}
	uc := newTestUC(repo)

	err := uc.Delete(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("repo.Delete was not called")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return nil, sql.ErrNoRows
		},
	}
	uc := newTestUC(repo)

	err := uc.Delete(context.Background(), 999, 10)
	if !errors.Is(err, ErrPOSNotFound) {
		t.Errorf("got %v, want ErrPOSNotFound", err)
	}
}

func TestDelete_WrongOrg(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
	}
	uc := newTestUC(repo)

	err := uc.Delete(context.Background(), 1, 999)
	if !errors.Is(err, ErrNotPOSOwner) {
		t.Errorf("got %v, want ErrNotPOSOwner", err)
	}
}

func TestDelete_RepoError(t *testing.T) {
	pos := testPOS(1, 10)
	repo := &mockPOSRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.POSLocation, error) {
			return pos, nil
		},
		deleteFn: func(_ context.Context, _ int) error {
			return errors.New("db error")
		},
	}
	uc := newTestUC(repo)

	err := uc.Delete(context.Background(), 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
