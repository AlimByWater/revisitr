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
	UpsertAggregate(ctx context.Context, agg *entity.IntegrationAggregate) error
	UpsertClientMap(ctx context.Context, mapping *entity.IntegrationClientMap) error
	MatchClients(ctx context.Context, integrationID int) (int, error)
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

// SyncAggregates syncs daily aggregates and client phone mappings from a POS provider.
func (s *SyncService) SyncAggregates(ctx context.Context, integration *entity.Integration) error {
	provider, err := NewProvider(integration)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}

	since := time.Now().Add(-30 * 24 * time.Hour)
	if integration.LastSyncAt != nil {
		since = *integration.LastSyncAt
	}

	aggregates, err := provider.GetDailyAggregates(ctx, since, time.Now())
	if err != nil {
		s.integrations.UpdateLastSync(ctx, integration.ID, "error")
		return fmt.Errorf("get daily aggregates: %w", err)
	}

	for _, agg := range aggregates {
		intAgg := &entity.IntegrationAggregate{
			IntegrationID: integration.ID,
			Date:          agg.Date,
			Revenue:       agg.Revenue,
			AvgCheck:      agg.AvgCheck,
			TxCount:       agg.TxCount,
			GuestCount:    agg.GuestCount,
		}
		if err := s.integrations.UpsertAggregate(ctx, intAgg); err != nil {
			s.logger.Error("upsert aggregate", "error", err, "date", agg.Date, "integration_id", integration.ID)
		}

		// Insert phone mappings for client matching
		for _, phone := range agg.Phones {
			mapping := &entity.IntegrationClientMap{
				IntegrationID: integration.ID,
				ExternalPhone: phone,
			}
			if err := s.integrations.UpsertClientMap(ctx, mapping); err != nil {
				s.logger.Error("upsert client map", "error", err, "phone", phone)
			}
		}
	}

	// Match phones to existing bot_clients
	matched, err := s.integrations.MatchClients(ctx, integration.ID)
	if err != nil {
		s.logger.Error("match clients", "error", err, "integration_id", integration.ID)
	} else if matched > 0 {
		s.logger.Info("matched clients", "integration_id", integration.ID, "count", matched)
	}

	s.integrations.UpdateLastSync(ctx, integration.ID, "active")
	s.logger.Info("aggregate sync completed", "integration_id", integration.ID, "days_synced", len(aggregates))
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
