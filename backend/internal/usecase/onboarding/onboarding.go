package onboarding

import (
	"context"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type onboardingRepo interface {
	GetByOrgID(ctx context.Context, orgID int) (*entity.Organization, error)
	UpdateState(ctx context.Context, orgID int, state entity.OnboardingState) error
	SetCompleted(ctx context.Context, orgID int, completed bool) error
}

type Usecase struct {
	logger *slog.Logger
	repo   onboardingRepo
}

func New(repo onboardingRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) GetState(ctx context.Context, orgID int) (*entity.Organization, error) {
	return uc.repo.GetByOrgID(ctx, orgID)
}

func (uc *Usecase) UpdateStep(ctx context.Context, orgID int, req entity.UpdateOnboardingRequest) (*entity.OnboardingState, error) {
	org, err := uc.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	state := org.OnboardingState
	if state.Steps == nil {
		state.Steps = make(map[string]OnboardingStep)
	}

	step := entity.OnboardingStep{
		Completed: req.Completed,
		Skipped:   req.Skipped,
		EntityID:  req.EntityID,
	}
	state.Steps[req.Step] = step

	// Advance current step
	completedCount := 0
	for _, s := range state.Steps {
		if s.Completed || s.Skipped {
			completedCount++
		}
	}
	state.CurrentStep = completedCount

	if err := uc.repo.UpdateState(ctx, orgID, state); err != nil {
		return nil, fmt.Errorf("update onboarding state: %w", err)
	}

	return &state, nil
}

func (uc *Usecase) Complete(ctx context.Context, orgID int) error {
	return uc.repo.SetCompleted(ctx, orgID, true)
}

func (uc *Usecase) Reset(ctx context.Context, orgID int) error {
	emptyState := entity.OnboardingState{
		CurrentStep: 1,
		Steps:       make(map[string]entity.OnboardingStep),
	}

	if err := uc.repo.UpdateState(ctx, orgID, emptyState); err != nil {
		return err
	}
	return uc.repo.SetCompleted(ctx, orgID, false)
}

type OnboardingStep = entity.OnboardingStep
