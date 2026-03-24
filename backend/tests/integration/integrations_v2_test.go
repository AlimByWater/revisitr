//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

// integrationResp and mustCreateIntegration defined in integrations_test.go

// --- Phase 2: Mock provider tests ---

func TestIntegration_TestConnection_Mock(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "MockOwner", "Mock Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	resp := doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/integrations/%d/test", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestIntegration_Sync_Mock(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "SyncMock", "Sync Mock Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	resp := doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/integrations/%d/sync", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// After sync, integration should be queryable for orders
	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(),
			"DELETE FROM external_orders WHERE integration_id = $1", intg.ID)
	})
}

func TestIntegration_GetOrders(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "OrdersOwner", "Orders Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	// Sync to populate orders
	resp := doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/integrations/%d/sync", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(),
			"DELETE FROM external_orders WHERE integration_id = $1", intg.ID)
	})

	// Get orders
	resp = doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d/orders?limit=10", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var result struct {
		Items []interface{} `json:"items"`
		Total int           `json:"total"`
	}
	decodeJSON(t, resp, &result)

	if result.Total == 0 {
		t.Error("expected some orders after sync")
	}
}

func TestIntegration_GetCustomers(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CustOwner", "Cust Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d/customers?limit=5", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var customers []interface{}
	decodeJSON(t, resp, &customers)

	if len(customers) == 0 {
		t.Error("expected mock customers")
	}
	if len(customers) > 5 {
		t.Errorf("expected at most 5 customers (limit), got %d", len(customers))
	}
}

func TestIntegration_GetMenu(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "MenuOwner", "Menu Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d/menu", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var menu struct {
		Categories []interface{} `json:"categories"`
	}
	decodeJSON(t, resp, &menu)

	if len(menu.Categories) == 0 {
		t.Error("expected mock menu categories")
	}
}

func TestIntegration_GetStats(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "StatsOwner", "Stats Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	// Sync first
	resp := doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/integrations/%d/sync", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(),
			"DELETE FROM external_orders WHERE integration_id = $1", intg.ID)
	})

	// Get stats
	resp = doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d/stats", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestIntegration_GetAggregates(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "AggOwner", "Agg Org")
	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "mock")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d/aggregates", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}
