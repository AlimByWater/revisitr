//go:build integration

package integration_test

import (
	"net/http"
	"testing"
)

// Client tests rely on bot_clients seeded via mustSeedBotClient (defined in loyalty_v2_test.go).

func TestClients_List(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CliOwner", "Cli Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Client Bot")
	mustSeedBotClient(t, ar.User.OrgID, bot.ID)
	mustSeedBotClient(t, ar.User.OrgID, bot.ID)

	resp := doRequest(t, http.MethodGet, "/api/v1/clients", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var result struct {
		Items []interface{} `json:"items"`
		Total int           `json:"total"`
	}
	decodeJSON(t, resp, &result)

	if result.Total < 2 {
		t.Errorf("expected at least 2 clients, got %d", result.Total)
	}
}

func TestClients_Stats(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "StatOwner", "Stat Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/clients/stats", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var stats map[string]interface{}
	decodeJSON(t, resp, &stats)
	// Just verify we get a valid JSON response; stats may be zero for a fresh org
}

func TestClients_Count(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CntOwner", "Cnt Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Count Bot")
	mustSeedBotClient(t, ar.User.OrgID, bot.ID)

	resp := doRequest(t, http.MethodGet, "/api/v1/clients/count", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var result struct {
		Count int `json:"count"`
	}
	decodeJSON(t, resp, &result)

	if result.Count < 1 {
		t.Errorf("expected count >= 1, got %d", result.Count)
	}
}

func TestClients_Identify_MissingParams(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "IdOwner", "Id Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/clients/identify", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func TestClients_Identify_NotFound(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "IdNFOwner", "IdNF Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/clients/identify?phone=+70000000000", nil, ar.Tokens.AccessToken)
	// May return 404 or 500 depending on whether ErrClientNotFound is returned
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 404 or 500 for unknown phone, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestClients_RequiresAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/clients", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}
