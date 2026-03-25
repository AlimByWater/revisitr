package onboarding

import (
	"context"
	"testing"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockOnboardingRepo struct {
	getByOrgIDFn   func(ctx context.Context, orgID int) (*entity.Organization, error)
	updateStateFn  func(ctx context.Context, orgID int, state entity.OnboardingState) error
	setCompletedFn func(ctx context.Context, orgID int, completed bool) error
}

func (m *mockOnboardingRepo) GetByOrgID(ctx context.Context, orgID int) (*entity.Organization, error) {
	if m.getByOrgIDFn != nil {
		return m.getByOrgIDFn(ctx, orgID)
	}
	return nil, nil
}
func (m *mockOnboardingRepo) UpdateState(ctx context.Context, orgID int, state entity.OnboardingState) error {
	if m.updateStateFn != nil {
		return m.updateStateFn(ctx, orgID, state)
	}
	return nil
}
func (m *mockOnboardingRepo) SetCompleted(ctx context.Context, orgID int, completed bool) error {
	if m.setCompletedFn != nil {
		return m.setCompletedFn(ctx, orgID, completed)
	}
	return nil
}

// --- tests ---

func TestGetState_NewOrg(t *testing.T) {
	repo := &mockOnboardingRepo{
		getByOrgIDFn: func(_ context.Context, _ int) (*entity.Organization, error) {
			return &entity.Organization{
				ID:                  1,
				Name:                "Test Org",
				OnboardingCompleted: false,
				OnboardingState:     entity.OnboardingState{},
			}, nil
		},
	}
	uc := New(repo)

	org, err := uc.GetState(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.OnboardingCompleted {
		t.Error("expected onboarding_completed=false for new org")
	}
	if org.OnboardingState.CurrentStep != 0 {
		t.Errorf("expected current_step=0, got %d", org.OnboardingState.CurrentStep)
	}
}

func TestGetState_ExistingState(t *testing.T) {
	repo := &mockOnboardingRepo{
		getByOrgIDFn: func(_ context.Context, _ int) (*entity.Organization, error) {
			return &entity.Organization{
				ID:                  1,
				Name:                "Existing Org",
				OnboardingCompleted: false,
				OnboardingState: entity.OnboardingState{
					CurrentStep: 3,
					Steps: map[string]entity.OnboardingStep{
						"1": {Completed: true},
						"2": {Completed: true},
					},
				},
			}, nil
		},
	}
	uc := New(repo)

	org, err := uc.GetState(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.OnboardingState.CurrentStep != 3 {
		t.Errorf("expected current_step=3, got %d", org.OnboardingState.CurrentStep)
	}
	if len(org.OnboardingState.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(org.OnboardingState.Steps))
	}
}

func TestUpdateStep(t *testing.T) {
	var savedState entity.OnboardingState
	repo := &mockOnboardingRepo{
		getByOrgIDFn: func(_ context.Context, _ int) (*entity.Organization, error) {
			return &entity.Organization{
				ID:   1,
				Name: "Test Org",
				OnboardingState: entity.OnboardingState{
					CurrentStep: 1,
					Steps:       map[string]entity.OnboardingStep{},
				},
			}, nil
		},
		updateStateFn: func(_ context.Context, _ int, state entity.OnboardingState) error {
			savedState = state
			return nil
		},
	}
	uc := New(repo)

	entityID := 42
	result, err := uc.UpdateStep(context.Background(), 1, entity.UpdateOnboardingRequest{
		Step:      1,
		Completed: true,
		EntityID:  &entityID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CurrentStep != 2 {
		t.Errorf("expected current_step advanced to 2, got %d", result.CurrentStep)
	}
	step, ok := result.Steps["1"]
	if !ok {
		t.Fatal("expected step '1' to exist")
	}
	if !step.Completed {
		t.Error("expected step 1 to be completed")
	}
	if step.EntityID == nil || *step.EntityID != 42 {
		t.Errorf("expected entity_id=42, got %v", step.EntityID)
	}
	// verify the state was saved
	if savedState.CurrentStep != 2 {
		t.Errorf("expected saved state current_step=2, got %d", savedState.CurrentStep)
	}
}

func TestComplete(t *testing.T) {
	var completedValue bool
	repo := &mockOnboardingRepo{
		setCompletedFn: func(_ context.Context, _ int, completed bool) error {
			completedValue = completed
			return nil
		},
	}
	uc := New(repo)

	if err := uc.Complete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !completedValue {
		t.Error("expected SetCompleted called with true")
	}
}

func TestReset(t *testing.T) {
	var savedState entity.OnboardingState
	var completedValue bool
	repo := &mockOnboardingRepo{
		updateStateFn: func(_ context.Context, _ int, state entity.OnboardingState) error {
			savedState = state
			return nil
		},
		setCompletedFn: func(_ context.Context, _ int, completed bool) error {
			completedValue = completed
			return nil
		},
	}
	uc := New(repo)

	if err := uc.Reset(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedState.CurrentStep != 1 {
		t.Errorf("expected reset current_step=1, got %d", savedState.CurrentStep)
	}
	if len(savedState.Steps) != 0 {
		t.Errorf("expected empty steps after reset, got %d", len(savedState.Steps))
	}
	if completedValue {
		t.Error("expected SetCompleted called with false")
	}
}
