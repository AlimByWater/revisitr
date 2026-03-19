package pos

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type integrationsRepo interface {
	UpdateLastSync(ctx context.Context, id int, status string) error
	UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error
	GetByID(ctx context.Context, id int) (*entity.Integration, error)
	GetActive(ctx context.Context) ([]entity.Integration, error)
}

type clientsRepo interface {
	GetByPhone(ctx context.Context, orgID int, phone string) (*entity.BotClient, error)
}

// SyncService orchestrates POS data synchronisation using provider adapters.
type SyncService struct {
	integrations integrationsRepo
	clients      clientsRepo
	logger       *slog.Logger
}

func NewSyncService(integrations integrationsRepo, clients clientsRepo, logger *slog.Logger) *SyncService {
	return &SyncService{
		integrations: integrations,
		clients:      clients,
		logger:       logger,
	}
}

// TestConnection tests connectivity for an integration.
func (s *SyncService) TestConnection(ctx context.Context, integration *entity.Integration) error {
	provider, err := NewProvider(integration)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	return provider.TestConnection(ctx)
}

// GetCustomers returns customers from the POS provider.
func (s *SyncService) GetCustomers(ctx context.Context, integration *entity.Integration, opts CustomerListOpts) ([]POSCustomer, error) {
	provider, err := NewProvider(integration)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}
	return provider.GetCustomers(ctx, opts)
}

// GetMenu returns the menu from the POS provider.
func (s *SyncService) GetMenu(ctx context.Context, integration *entity.Integration) (*POSMenu, error) {
	provider, err := NewProvider(integration)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}
	return provider.GetMenu(ctx)
}

// Sync syncs orders from a single integration.
func (s *SyncService) Sync(ctx context.Context, integration *entity.Integration) error {
	provider, err := NewProvider(integration)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}

	if err := provider.TestConnection(ctx); err != nil {
		s.integrations.UpdateLastSync(ctx, integration.ID, "error")
		return fmt.Errorf("test connection: %w", err)
	}

	since := time.Now().Add(-30 * 24 * time.Hour) // initial sync: 30 days
	if integration.LastSyncAt != nil {
		since = *integration.LastSyncAt
	}

	orders, err := provider.GetOrders(ctx, since, time.Now())
	if err != nil {
		s.integrations.UpdateLastSync(ctx, integration.ID, "error")
		return fmt.Errorf("get orders: %w", err)
	}

	for _, order := range orders {
		extOrder := &entity.ExternalOrder{
			IntegrationID: integration.ID,
			ExternalID:    order.ExternalID,
			Items:         toEntityItems(order.Items),
			Total:         order.Total,
			OrderedAt:     &order.OrderedAt,
		}

		if order.CustomerPhone != "" && s.clients != nil {
			if client, err := s.clients.GetByPhone(ctx, integration.OrgID, order.CustomerPhone); err == nil {
				extOrder.ClientID = &client.ID
			}
		}

		if err := s.integrations.UpsertOrder(ctx, extOrder); err != nil {
			s.logger.Error("upsert order", "error", err, "external_id", order.ExternalID)
		}
	}

	s.integrations.UpdateLastSync(ctx, integration.ID, "active")
	s.logger.Info("sync completed", "integration_id", integration.ID, "orders_synced", len(orders))
	return nil
}

// SyncAll syncs all active integrations (called by scheduler).
func (s *SyncService) SyncAll(ctx context.Context) error {
	intgs, err := s.integrations.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("get active integrations: %w", err)
	}

	for i := range intgs {
		if err := s.Sync(ctx, &intgs[i]); err != nil {
			s.logger.Error("sync failed", "integration_id", intgs[i].ID, "type", intgs[i].Type, "error", err)
		}
	}
	return nil
}

func toEntityItems(items []POSOrderItem) entity.OrderItems {
	result := make(entity.OrderItems, len(items))
	for i, it := range items {
		result[i] = entity.OrderItem{
			Name:     it.Name,
			Quantity: it.Quantity,
			Price:    it.Price,
		}
	}
	return result
}
