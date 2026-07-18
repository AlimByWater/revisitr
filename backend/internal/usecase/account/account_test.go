package account

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"revisitr/internal/entity"
)

type mockOrgsRepo struct {
	getByIDFn        func(ctx context.Context, orgID int) (*entity.Organization, error)
	updateTimezoneFn func(ctx context.Context, orgID int, timezone string) error
}

func (m *mockOrgsRepo) GetByID(ctx context.Context, orgID int) (*entity.Organization, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, orgID)
	}
	return &entity.Organization{ID: orgID, Timezone: "Europe/Moscow"}, nil
}
func (m *mockOrgsRepo) UpdateTimezone(ctx context.Context, orgID int, timezone string) error {
	if m.updateTimezoneFn != nil {
		return m.updateTimezoneFn(ctx, orgID, timezone)
	}
	return nil
}

func newTestUsecase(t *testing.T, repo *mockOrgsRepo) *Usecase {
	t.Helper()
	if repo == nil {
		repo = &mockOrgsRepo{}
	}
	uc := New(repo)
	if err := uc.Init(context.Background(), slog.Default()); err != nil {
		t.Fatalf("init usecase: %v", err)
	}
	return uc
}

func ptr[T any](value T) *T { return &value }

func TestUpdateOrganizationInvalidTimezone(t *testing.T) {
	uc := newTestUsecase(t, nil)
	_, err := uc.UpdateOrganization(context.Background(), 1,
		entity.UpdateOrganizationRequest{Timezone: ptr("Mars/Olympus")})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestUpdateOrganizationTimezone(t *testing.T) {
	var savedTZ string
	repo := &mockOrgsRepo{
		updateTimezoneFn: func(_ context.Context, _ int, timezone string) error {
			savedTZ = timezone
			return nil
		},
	}
	uc := newTestUsecase(t, repo)
	org, err := uc.UpdateOrganization(context.Background(), 1,
		entity.UpdateOrganizationRequest{Timezone: ptr("Asia/Yekaterinburg")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedTZ != "Asia/Yekaterinburg" {
		t.Errorf("saved timezone %q, want Asia/Yekaterinburg", savedTZ)
	}
	if org == nil {
		t.Error("expected updated organization, got nil")
	}
}

func TestGetOrganizationMissing(t *testing.T) {
	repo := &mockOrgsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Organization, error) { return nil, nil },
	}
	uc := newTestUsecase(t, repo)
	if _, err := uc.GetOrganization(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
