package pos

import (
	"context"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

// RKeeperProvider connects to r-keeper 7 XML interface.
// Placeholder — will be implemented in Phase C.
type RKeeperProvider struct {
	baseURL  string
	username string
	password string
}

func NewRKeeperProvider(cfg entity.IntegrationConfig) (*RKeeperProvider, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("rkeeper: api_url is required")
	}
	return &RKeeperProvider{
		baseURL:  cfg.APIURL,
		username: cfg.Username,
		password: cfg.Password,
	}, nil
}

func (p *RKeeperProvider) TestConnection(_ context.Context) error {
	return fmt.Errorf("r-keeper provider not yet implemented")
}

func (p *RKeeperProvider) GetCustomers(_ context.Context, _ CustomerListOpts) ([]POSCustomer, error) {
	return nil, fmt.Errorf("r-keeper provider not yet implemented")
}

func (p *RKeeperProvider) GetOrders(_ context.Context, _, _ time.Time) ([]POSOrder, error) {
	return nil, fmt.Errorf("r-keeper provider not yet implemented")
}

func (p *RKeeperProvider) GetMenu(_ context.Context) (*POSMenu, error) {
	return nil, fmt.Errorf("r-keeper provider not yet implemented")
}

func (p *RKeeperProvider) GetDailyAggregates(_ context.Context, _, _ time.Time) ([]POSDailyAggregate, error) {
	return nil, fmt.Errorf("r-keeper provider not yet implemented")
}
