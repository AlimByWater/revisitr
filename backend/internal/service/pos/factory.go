package pos

import (
	"fmt"

	"revisitr/internal/entity"
)

// NewProvider creates a POSProvider based on integration type.
func NewProvider(integration *entity.Integration) (POSProvider, error) {
	switch integration.Type {
	case "iiko":
		return NewIikoProvider(integration.Config)
	case "rkeeper":
		return NewRKeeperProvider(integration.Config)
	case "mock":
		return NewMockProvider()
	default:
		return nil, fmt.Errorf("unknown integration type: %s", integration.Type)
	}
}
