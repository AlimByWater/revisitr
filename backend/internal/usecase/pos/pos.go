package pos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrPOSNotFound = fmt.Errorf("pos location not found")
	ErrNotPOSOwner = fmt.Errorf("not authorized to manage this location")
)

type posRepo interface {
	Create(ctx context.Context, pos *entity.POSLocation) error
	GetByID(ctx context.Context, id int) (*entity.POSLocation, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.POSLocation, error)
	Update(ctx context.Context, pos *entity.POSLocation) error
	Delete(ctx context.Context, id int) error
}

type Usecase struct {
	logger  *slog.Logger
	posRepo posRepo
}

func New(posRepo posRepo) *Usecase {
	return &Usecase{
		posRepo: posRepo,
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Create(ctx context.Context, orgID int, req *entity.CreatePOSRequest) (*entity.POSLocation, error) {
	pos := &entity.POSLocation{
		OrgID:    orgID,
		Name:     req.Name,
		Address:  req.Address,
		Phone:    req.Phone,
		Schedule: req.Schedule,
		IsActive: true,
	}

	if pos.Schedule == nil {
		pos.Schedule = entity.Schedule{}
	}

	if err := uc.posRepo.Create(ctx, pos); err != nil {
		return nil, fmt.Errorf("pos.Create: %w", err)
	}

	return pos, nil
}

func (uc *Usecase) GetByOrgID(ctx context.Context, orgID int) ([]entity.POSLocation, error) {
	locations, err := uc.posRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("pos.GetByOrgID: %w", err)
	}
	return locations, nil
}

func (uc *Usecase) GetByID(ctx context.Context, id, orgID int) (*entity.POSLocation, error) {
	pos, err := uc.posRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPOSNotFound
		}
		return nil, fmt.Errorf("pos.GetByID: %w", err)
	}

	if pos.OrgID != orgID {
		return nil, ErrNotPOSOwner
	}

	return pos, nil
}

func (uc *Usecase) Update(ctx context.Context, id, orgID int, req *entity.UpdatePOSRequest) (*entity.POSLocation, error) {
	pos, err := uc.posRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPOSNotFound
		}
		return nil, fmt.Errorf("pos.Update get: %w", err)
	}

	if pos.OrgID != orgID {
		return nil, ErrNotPOSOwner
	}

	if req.Name != nil {
		pos.Name = *req.Name
	}
	if req.Address != nil {
		pos.Address = *req.Address
	}
	if req.Phone != nil {
		pos.Phone = *req.Phone
	}
	if req.Schedule != nil {
		pos.Schedule = *req.Schedule
	}
	if req.IsActive != nil {
		pos.IsActive = *req.IsActive
	}

	if err := uc.posRepo.Update(ctx, pos); err != nil {
		return nil, fmt.Errorf("pos.Update: %w", err)
	}

	// Re-fetch to get updated_at from DB
	pos, err = uc.posRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("pos.Update refetch: %w", err)
	}

	return pos, nil
}

func (uc *Usecase) Delete(ctx context.Context, id, orgID int) error {
	pos, err := uc.posRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPOSNotFound
		}
		return fmt.Errorf("pos.Delete get: %w", err)
	}

	if pos.OrgID != orgID {
		return ErrNotPOSOwner
	}

	if err := uc.posRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("pos.Delete: %w", err)
	}

	return nil
}
