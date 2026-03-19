package integrations

import "errors"

var (
	ErrIntegrationNotFound = errors.New("integration not found")
	ErrNotIntegrationOwner = errors.New("not authorized")
)
