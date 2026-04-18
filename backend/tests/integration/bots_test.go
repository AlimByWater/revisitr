//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

// botResp mirrors the JSON shape returned by the bots API.
type botResp struct {
	ID     int    `json:"id"`
	OrgID  int    `json:"org_id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// mustCreateBot creates a bot via the API and registers cleanup to delete it.
func mustCreateBot(t *testing.T, token string, name string) botResp {
	t.Helper()

	body := map[string]string{
		"name":  name,
		"token": fmt.Sprintf("fake-tg-token-%d", tokenCounter()),
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/bots", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var bot botResp
	decodeJSON(t, resp, &bot)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(
			t.Context(),
			"DELETE FROM bots WHERE id = $1", bot.ID,
		)
	})

	return bot
}

var counter int

func tokenCounter() int {
	counter++
	return counter
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestBots_Create(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")

	bot := mustCreateBot(t, ar.Tokens.AccessToken, "My Test Bot")

	if bot.ID == 0 {
		t.Fatal("expected non-zero bot ID")
	}
	if bot.Name != "My Test Bot" {
		t.Fatalf("expected name %q, got %q", "My Test Bot", bot.Name)
	}
	if bot.Status != "pending" {
		t.Fatalf("expected status %q, got %q", "pending", bot.Status)
	}
	if bot.OrgID != ar.User.OrgID {
		t.Fatalf("expected org_id %d, got %d", ar.User.OrgID, bot.OrgID)
	}
}

func TestBots_Create_RequiresAuth(t *testing.T) {
	body := map[string]string{"name": "Unauthorized Bot", "token": "fake-token"}
	resp := doRequest(t, http.MethodPost, "/api/v1/bots", body, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestBots_List(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")

	mustCreateBot(t, ar.Tokens.AccessToken, "Bot Alpha")
	mustCreateBot(t, ar.Tokens.AccessToken, "Bot Beta")

	resp := doRequest(t, http.MethodGet, "/api/v1/bots", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var bots []botResp
	decodeJSON(t, resp, &bots)

	if len(bots) < 2 {
		t.Fatalf("expected at least 2 bots, got %d", len(bots))
	}
}

func TestBots_List_IsolatedByOrg(t *testing.T) {
	// Two separate users each create a bot — they must not see each other's bots.
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "Owner One", "Org One")
	mustCreateBot(t, ar1.Tokens.AccessToken, "Org1 Bot")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "Owner Two", "Org Two")
	mustCreateBot(t, ar2.Tokens.AccessToken, "Org2 Bot")

	resp1 := doRequest(t, http.MethodGet, "/api/v1/bots", nil, ar1.Tokens.AccessToken)
	assertStatus(t, resp1, http.StatusOK)
	var bots1 []botResp
	decodeJSON(t, resp1, &bots1)
	for _, b := range bots1 {
		if b.OrgID != ar1.User.OrgID {
			t.Errorf("user1 sees bot from org %d (expected org %d)", b.OrgID, ar1.User.OrgID)
		}
	}

	resp2 := doRequest(t, http.MethodGet, "/api/v1/bots", nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp2, http.StatusOK)
	var bots2 []botResp
	decodeJSON(t, resp2, &bots2)
	for _, b := range bots2 {
		if b.OrgID != ar2.User.OrgID {
			t.Errorf("user2 sees bot from org %d (expected org %d)", b.OrgID, ar2.User.OrgID)
		}
	}
}

func TestBots_Get(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Fetch Me Bot")

	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/bots/%d", bot.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var fetched botResp
	decodeJSON(t, resp, &fetched)

	if fetched.ID != bot.ID {
		t.Fatalf("expected ID %d, got %d", bot.ID, fetched.ID)
	}
	if fetched.Name != bot.Name {
		t.Fatalf("expected name %q, got %q", bot.Name, fetched.Name)
	}
}

func TestBots_Get_NotFound(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/bots/999999999", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

func TestBots_Get_CrossOrgForbidden(t *testing.T) {
	// User 1 creates a bot; User 2 must not be able to fetch it.
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "Owner One", "Org One")
	bot := mustCreateBot(t, ar1.Tokens.AccessToken, "Private Bot")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "Owner Two", "Org Two")

	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/bots/%d", bot.ID), nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}

func TestBots_Update(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Original Name")

	newName := "Updated Name"
	body := map[string]string{"name": newName}
	resp := doRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/bots/%d", bot.ID), body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var updated botResp
	decodeJSON(t, resp, &updated)

	if updated.Name != newName {
		t.Fatalf("expected name %q after update, got %q", newName, updated.Name)
	}
}

func TestBots_Update_CrossOrgForbidden(t *testing.T) {
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "Owner One", "Org One")
	bot := mustCreateBot(t, ar1.Tokens.AccessToken, "Protected Bot")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "Owner Two", "Org Two")

	body := map[string]string{"name": "Hijacked Name"}
	resp := doRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/bots/%d", bot.ID), body, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}

func TestBots_Delete(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")

	// Create the bot manually so we can delete it ourselves (without t.Cleanup)
	body := map[string]string{"name": "To Delete", "token": fmt.Sprintf("fake-%d", tokenCounter())}
	createResp := doRequest(t, http.MethodPost, "/api/v1/bots", body, ar.Tokens.AccessToken)
	assertStatus(t, createResp, http.StatusCreated)
	var bot botResp
	decodeJSON(t, createResp, &bot)

	// Delete it
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/bots/%d", bot.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	// Verify it's gone
	getResp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/bots/%d", bot.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, getResp, http.StatusNotFound)
	getResp.Body.Close()
}

func TestBots_Delete_CrossOrgForbidden(t *testing.T) {
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "Owner One", "Org One")
	bot := mustCreateBot(t, ar1.Tokens.AccessToken, "Safe Bot")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "Owner Two", "Org Two")

	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/bots/%d", bot.ID), nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}

func TestBots_GetSettings(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Settings Bot")

	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/bots/%d/settings", bot.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var settings map[string]interface{}
	decodeJSON(t, resp, &settings)

	// New bot should have empty arrays for modules, buttons, registration_form
	if _, ok := settings["modules"]; !ok {
		t.Error("expected 'modules' field in settings")
	}
	if _, ok := settings["buttons"]; !ok {
		t.Error("expected 'buttons' field in settings")
	}
}

func TestBots_UpdateSettings(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Bot Owner", "Bot Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Settings Bot")

	body := map[string]interface{}{
		"modules":         []string{"loyalty", "pos"},
		"welcome_message": "Hello from integration test!",
	}
	resp := doRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/bots/%d/settings", bot.ID), body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify persisted
	getResp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/bots/%d/settings", bot.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, getResp, http.StatusOK)

	var settings map[string]interface{}
	decodeJSON(t, getResp, &settings)

	modules, _ := settings["modules"].([]interface{})
	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}
	if settings["welcome_message"] != "Hello from integration test!" {
		t.Fatalf("welcome_message not persisted: %v", settings["welcome_message"])
	}
}
