package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

var (
	ErrNotFound   = errors.New("organization not found")
	ErrValidation = errors.New("validation failed")
)

type orgsRepo interface {
	GetByID(ctx context.Context, orgID int) (*entity.Organization, error)
	UpdateTimezone(ctx context.Context, orgID int, timezone string) error
}

type Usecase struct {
	logger *slog.Logger
	repo   orgsRepo
}

func New(repo orgsRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) GetOrganization(ctx context.Context, orgID int) (*entity.Organization, error) {
	org, err := uc.repo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrNotFound
	}
	return org, nil
}

func (uc *Usecase) UpdateOrganization(ctx context.Context, orgID int, req entity.UpdateOrganizationRequest) (*entity.Organization, error) {
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			return nil, fmt.Errorf("%w: unknown timezone %q", ErrValidation, *req.Timezone)
		}
		if err := uc.repo.UpdateTimezone(ctx, orgID, *req.Timezone); err != nil {
			return nil, err
		}
	}
	return uc.GetOrganization(ctx, orgID)
}
