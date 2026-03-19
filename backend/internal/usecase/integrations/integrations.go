package integrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type integrationsRepo interface {
	Create(ctx context.Context, intg *entity.Integration) error
	GetByID(ctx context.Context, id int) (*entity.Integration, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Integration, error)
	Update(ctx context.Context, intg *entity.Integration) error
	Delete(ctx context.Context, id int) error
	UpdateLastSync(ctx context.Context, id int, status string) error
	UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error
}

type syncService interface {
	Sync(ctx context.Context, integration *entity.Integration) error
}

type Usecase struct {
	logger       *slog.Logger
	integrations integrationsRepo
	syncSvc      syncService
}

func New(integrations integrationsRepo, syncSvc syncService) *Usecase {
	return &Usecase{integrations: integrations, syncSvc: syncSvc}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Create(ctx context.Context, orgID int, req *entity.CreateIntegrationRequest) (*entity.Integration, error) {
	intg := &entity.Integration{
		OrgID:  orgID,
		Type:   req.Type,
		Config: req.Config,
		Status: "inactive",
	}

	if err := uc.integrations.Create(ctx, intg); err != nil {
		return nil, fmt.Errorf("create integration: %w", err)
	}

	return intg, nil
}

func (uc *Usecase) List(ctx context.Context, orgID int) ([]entity.Integration, error) {
	intgs, err := uc.integrations.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list integrations: %w", err)
	}
	return intgs, nil
}

func (uc *Usecase) GetByID(ctx context.Context, id, orgID int) (*entity.Integration, error) {
	intg, err := uc.integrations.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIntegrationNotFound
		}
		return nil, fmt.Errorf("get integration: %w", err)
	}

	if intg.OrgID != orgID {
		return nil, ErrNotIntegrationOwner
	}

	return intg, nil
}

func (uc *Usecase) Update(ctx context.Context, id, orgID int, req *entity.UpdateIntegrationRequest) (*entity.Integration, error) {
	intg, err := uc.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Config != nil {
		intg.Config = *req.Config
	}
	if req.Status != nil {
		intg.Status = *req.Status
	}

	if err := uc.integrations.Update(ctx, intg); err != nil {
		return nil, fmt.Errorf("update integration: %w", err)
	}

	return intg, nil
}

func (uc *Usecase) Delete(ctx context.Context, id, orgID int) error {
	if _, err := uc.GetByID(ctx, id, orgID); err != nil {
		return err
	}

	if err := uc.integrations.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete integration: %w", err)
	}

	return nil
}

func (uc *Usecase) SyncNow(ctx context.Context, id, orgID int) error {
	intg, err := uc.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}

	if err := uc.syncSvc.Sync(ctx, intg); err != nil {
		if updateErr := uc.integrations.UpdateLastSync(ctx, id, "error"); updateErr != nil {
			uc.logger.Error("update integration status after sync error", "error", updateErr, "id", id)
		}
		return fmt.Errorf("sync integration: %w", err)
	}

	return nil
}
