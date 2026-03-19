package pos

import (
	"context"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

// IikoProvider connects to the iiko Cloud API.
// Placeholder — will be implemented in Phase B.
type IikoProvider struct {
	baseURL  string
	apiLogin string
	orgID    string
}

func NewIikoProvider(cfg entity.IntegrationConfig) (*IikoProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("iiko: api_key (API Login) is required")
	}
	baseURL := cfg.APIURL
	if baseURL == "" {
		baseURL = "https://api-ru.iiko.services/api/1"
	}
	return &IikoProvider{
		baseURL:  baseURL,
		apiLogin: cfg.APIKey,
		orgID:    cfg.OrgID,
	}, nil
}

func (p *IikoProvider) TestConnection(_ context.Context) error {
	return fmt.Errorf("iiko provider not yet implemented")
}

func (p *IikoProvider) GetCustomers(_ context.Context, _ CustomerListOpts) ([]POSCustomer, error) {
	return nil, fmt.Errorf("iiko provider not yet implemented")
}

func (p *IikoProvider) GetOrders(_ context.Context, _, _ time.Time) ([]POSOrder, error) {
	return nil, fmt.Errorf("iiko provider not yet implemented")
}

func (p *IikoProvider) GetMenu(_ context.Context) (*POSMenu, error) {
	return nil, fmt.Errorf("iiko provider not yet implemented")
}
