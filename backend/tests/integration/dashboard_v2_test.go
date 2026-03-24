//go:build integration

package integration_test

import (
	"net/http"
	"testing"
	"time"
)

func TestDashboard_Widgets(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "DashOwner", "Dash Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/dashboard/widgets", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestDashboard_Charts(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ChartOwner", "Chart Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/dashboard/charts", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestDashboard_Sales(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "SalesOwner", "Sales Org")

	from := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")

	resp := doRequest(t, http.MethodGet,
		"/api/v1/dashboard/sales?from="+from+"&to="+to,
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestDashboard_RequiresAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/dashboard/widgets", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}
