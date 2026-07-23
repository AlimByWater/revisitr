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

type menusRepo interface {
	GetByOrgAndIntegration(ctx context.Context, orgID, integrationID int) (*entity.Menu, error)
	GetFullMenu(ctx context.Context, menuID int) (*entity.Menu, error)
	UpsertFromPOS(ctx context.Context, integrationID, orgID int, posMenu *POSMenu) (SyncMenuResult, error)
}

// SyncService orchestrates POS data synchronisation using provider adapters.
type SyncService struct {
	integrations integrationsRepo
	clients      clientsRepo
	menus        menusRepo
	logger       *slog.Logger
}

func NewSyncService(integrations integrationsRepo, clients clientsRepo, logger *slog.Logger, opts ...SyncOption) *SyncService {
	s := &SyncService{
		integrations: integrations,
		clients:      clients,
		logger:       logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SyncOption configures optional dependencies for SyncService.
type SyncOption func(*SyncService)

// WithMenus enables menu auto-import during POS sync.
func WithMenus(menus menusRepo) SyncOption {
	return func(s *SyncService) {
		s.menus = menus
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

// discoveryProvider is implemented by providers that can list selectable
// resources (organizations, menus) before an integration is fully configured.
type discoveryProvider interface {
	ListOrganizations(ctx context.Context) ([]POSOrganization, error)
	ListExternalMenus(ctx context.Context) ([]POSExternalMenu, error)
}

// Discover fetches selectable resources (organizations + external menus) for a
// not-yet-saved integration, so onboarding can offer pick-lists. Only the API
// credentials in cfg are required; org_id is what the user is choosing.
func (s *SyncService) Discover(ctx context.Context, integrationType string, cfg entity.IntegrationConfig) (*POSDiscovery, error) {
	provider, err := newDiscoveryProvider(integrationType, cfg)
	if err != nil {
		return nil, err
	}

	orgs, err := provider.ListOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}

	menus, err := provider.ListExternalMenus(ctx)
	if err != nil {
		// Menu listing is optional; surface orgs even if menus are unavailable.
		s.logger.Warn("list external menus", "error", err)
		menus = nil
	}

	return &POSDiscovery{Organizations: orgs, ExternalMenus: menus}, nil
}

// newDiscoveryProvider builds a provider for discovery, where org_id is not yet
// known. cfg gets a placeholder org_id so construction validation passes; the
// discovery calls themselves do not use org_id.
func newDiscoveryProvider(integrationType string, cfg entity.IntegrationConfig) (discoveryProvider, error) {
	switch integrationType {
	case "iiko":
		if cfg.OrgID == "" {
			cfg.OrgID = "discovery"
		}
		return NewIikoProvider(cfg)
	default:
		return nil, fmt.Errorf("discovery not supported for integration type: %s", integrationType)
	}
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

func (s *SyncService) GetImportedMenu(ctx context.Context, integration *entity.Integration) (*entity.Menu, error) {
	if s.menus == nil {
		return nil, nil
	}
	menu, err := s.menus.GetByOrgAndIntegration(ctx, integration.OrgID, integration.ID)
	if err != nil || menu == nil {
		return menu, err
	}
	return s.menus.GetFullMenu(ctx, menu.ID)
}

// Sync syncs orders and menu data from a single integration.
func (s *SyncService) Sync(ctx context.Context, integration *entity.Integration) (*SyncResult, error) {
	provider, err := NewProvider(integration)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	if err := provider.TestConnection(ctx); err != nil {
		s.integrations.UpdateLastSync(ctx, integration.ID, "error")
		return nil, fmt.Errorf("test connection: %w", err)
	}

	since := time.Now().Add(-30 * 24 * time.Hour)
	if integration.LastSyncAt != nil {
		since = *integration.LastSyncAt
	}

	orders, err := provider.GetOrders(ctx, since, time.Now())
	if err != nil {
		s.integrations.UpdateLastSync(ctx, integration.ID, "error")
		return nil, fmt.Errorf("get orders: %w", err)
	}

	for _, order := range orders {
		extOrder := &entity.ExternalOrder{
			IntegrationID: integration.ID,
			ExternalID:    order.ExternalID,
			Source:        "delivery",
			Items:         toEntityItems(order.Items),
			Total:         order.Total,
			OrderedAt:     &order.OrderedAt,
		}
		if order.CustomerPhone != "" {
			phone := order.CustomerPhone
			extOrder.CustomerPhone = &phone
			if s.clients != nil {
				if client, err := s.clients.GetByPhone(ctx, integration.OrgID, order.CustomerPhone); err == nil {
					extOrder.ClientID = &client.ID
				}
			}
		}
		if err := s.integrations.UpsertOrder(ctx, extOrder); err != nil {
			s.logger.Error("upsert order", "error", err, "external_id", order.ExternalID)
		}
	}

	s.integrations.UpdateLastSync(ctx, integration.ID, "active")
	result := &SyncResult{OrdersSynced: len(orders)}
	if s.menus != nil {
		menuResult, err := s.syncMenu(ctx, provider, integration)
		if err != nil {
			s.logger.Error("menu sync failed", "error", err, "integration_id", integration.ID)
		} else {
			result.Menu = menuResult
		}
	}

	s.logger.Info("sync completed", "integration_id", integration.ID, "orders_synced", result.OrdersSynced)
	return result, nil
}

// syncMenu fetches the menu from the POS provider and upserts it into the database.
func (s *SyncService) syncMenu(ctx context.Context, provider POSProvider, integration *entity.Integration) (SyncMenuResult, error) {
	posMenu, err := provider.GetMenu(ctx)
	if err != nil {
		return SyncMenuResult{}, fmt.Errorf("get menu: %w", err)
	}
	if posMenu == nil || len(posMenu.Categories) == 0 {
		return SyncMenuResult{}, nil
	}

	result, err := s.menus.UpsertFromPOS(ctx, integration.ID, integration.OrgID, posMenu)
	if err != nil {
		return SyncMenuResult{}, fmt.Errorf("upsert menu: %w", err)
	}

	s.logger.Info("menu synced", "integration_id", integration.ID, "categories", len(posMenu.Categories), "added", result.Added, "updated", result.Updated, "missing", result.Missing)
	return result, nil
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
		if _, err := s.Sync(ctx, &intgs[i]); err != nil {
			s.logger.Error("sync failed", "integration_id", intgs[i].ID, "type", intgs[i].Type, "error", err)
			continue
		}
		if err := s.SyncAggregates(ctx, &intgs[i]); err != nil {
			s.logger.Error("aggregate sync failed", "integration_id", intgs[i].ID, "type", intgs[i].Type, "error", err)
		}
	}
	return nil
}

func toEntityItems(items []POSOrderItem) entity.OrderItems {
	result := make(entity.OrderItems, len(items))
	for i, it := range items {
		result[i] = entity.OrderItem{
			ExternalID: it.ExternalID,
			Name:       it.Name,
			Quantity:   it.Quantity,
			Price:      it.Price,
			Category:   it.Category,
		}
	}
	return result
}
