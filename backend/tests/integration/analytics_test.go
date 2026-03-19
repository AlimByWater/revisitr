//go:build integration

package integration_test

import (
	"net/http"
	"testing"
)

func TestAnalytics_Sales_RequiresAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/analytics/sales", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestAnalytics_Sales_OK(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Analyst", "Analytics Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/analytics/sales", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var body struct {
		Metrics map[string]interface{} `json:"metrics"`
	}
	decodeJSON(t, resp, &body)

	if body.Metrics == nil {
		t.Fatal("expected metrics object in sales response")
	}
	if _, ok := body.Metrics["total_amount"]; !ok {
		t.Error("expected total_amount in sales metrics")
	}
	if _, ok := body.Metrics["transaction_count"]; !ok {
		t.Error("expected transaction_count in sales metrics")
	}
}

func TestAnalytics_Sales_WithDateRange(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Analyst2", "Analytics Org2")

	resp := doRequest(t, http.MethodGet,
		"/api/v1/analytics/sales?from=2025-01-01&to=2025-12-31",
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestAnalytics_Sales_InvalidDateRange(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Analyst3", "Analytics Org3")

	// from > to should return 400
	resp := doRequest(t, http.MethodGet,
		"/api/v1/analytics/sales?from=2025-12-31&to=2025-01-01",
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func TestAnalytics_Loyalty_OK(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Analyst4", "Analytics Org4")

	resp := doRequest(t, http.MethodGet, "/api/v1/analytics/loyalty", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var body map[string]interface{}
	decodeJSON(t, resp, &body)

	if _, ok := body["new_clients"]; !ok {
		t.Error("expected new_clients in loyalty response")
	}
}

func TestAnalytics_Campaigns_OK(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Analyst5", "Analytics Org5")

	resp := doRequest(t, http.MethodGet, "/api/v1/analytics/campaigns", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}
