package pos

import (
	"context"
	"fmt"

	"revisitr/internal/entity"
)

type integrationsRepo interface {
	UpdateLastSync(ctx context.Context, id int, status string) error
	UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error
}

type clientsRepo interface {
	GetByPhone(ctx context.Context, orgID int, phone string) (*entity.BotClient, error)
}

// SyncService orchestrates POS data synchronisation.
// It is designed to be extended with real POS API clients in Wave 2.
type SyncService struct {
	integrations integrationsRepo
}

func NewSyncService(integrations integrationsRepo) *SyncService {
	return &SyncService{integrations: integrations}
}

// Sync syncs an integration. Currently a stub — real POS clients added in Wave 2.
func (s *SyncService) Sync(ctx context.Context, integration *entity.Integration) error {
	if err := s.integrations.UpdateLastSync(ctx, integration.ID, "active"); err != nil {
		return fmt.Errorf("pos.Sync update last sync: %w", err)
	}
	return nil
}

// SyncAll syncs all integrations for all orgs (called by scheduler).
func (s *SyncService) SyncAll(ctx context.Context) error {
	// Placeholder — will iterate over active integrations in Wave 2.
	return nil
}
