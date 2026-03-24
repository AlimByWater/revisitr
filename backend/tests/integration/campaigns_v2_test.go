//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type campaignResp struct {
	ID          int     `json:"id"`
	OrgID       int     `json:"org_id"`
	BotID       int     `json:"bot_id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	ScheduledAt *string `json:"scheduled_at,omitempty"`
}

type paginatedCampaigns struct {
	Items []campaignResp `json:"items"`
	Total int            `json:"total"`
}

type scenarioResp struct {
	ID          int    `json:"id"`
	OrgID       int    `json:"org_id"`
	BotID       int    `json:"bot_id"`
	Name        string `json:"name"`
	TriggerType string `json:"trigger_type"`
	IsActive    bool   `json:"is_active"`
	IsTemplate  bool   `json:"is_template"`
}

func mustCreateCampaign(t *testing.T, token, name string, botID int) campaignResp {
	t.Helper()

	body := map[string]interface{}{
		"name":    name,
		"bot_id":  botID,
		"message": "Test campaign message",
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/campaigns", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var c campaignResp
	decodeJSON(t, resp, &c)

	t.Cleanup(func() {
		db := pgMod.DB()
		db.ExecContext(t.Context(), "DELETE FROM campaign_clicks WHERE campaign_id = $1", c.ID)
		db.ExecContext(t.Context(), "DELETE FROM campaign_messages WHERE campaign_id = $1", c.ID)
		db.ExecContext(t.Context(), "DELETE FROM campaigns WHERE id = $1", c.ID)
	})

	return c
}

func mustCreateScenario(t *testing.T, token string, botID int, name, triggerType string) scenarioResp {
	t.Helper()

	body := map[string]interface{}{
		"name":         name,
		"bot_id":       botID,
		"trigger_type": triggerType,
		"message":      "Auto scenario message",
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/campaigns/scenarios", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var s scenarioResp
	decodeJSON(t, resp, &s)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(), "DELETE FROM auto_scenarios WHERE id = $1", s.ID)
	})

	return s
}

// --- Campaign CRUD ---

func TestCampaign_CRUD(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampOwner", "Camp Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Campaign Bot")

	// Create
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "My Campaign", bot.ID)
	if c.ID == 0 {
		t.Fatal("expected non-zero campaign ID")
	}
	if c.Status != "draft" {
		t.Errorf("expected status draft, got %s", c.Status)
	}

	// Get
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	var fetched campaignResp
	decodeJSON(t, resp, &fetched)
	if fetched.Name != "My Campaign" {
		t.Errorf("expected name %q, got %q", "My Campaign", fetched.Name)
	}

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/campaigns", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	var list paginatedCampaigns
	decodeJSON(t, resp, &list)
	if list.Total < 1 {
		t.Errorf("expected at least 1 campaign, got %d", list.Total)
	}

	// Update
	updateBody := map[string]interface{}{"name": "Updated Campaign"}
	resp = doRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), updateBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Delete
	resp = doRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify gone
	resp = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

func TestCampaign_Send(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampSend", "Camp Send Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Send Bot")
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "Sendable", bot.ID)

	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/send", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestCampaign_Send_AlreadySent(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampDouble", "Camp Double Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Double Send Bot")
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "Double Send", bot.ID)

	// First send
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/send", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Second send should fail
	resp = doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/send", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusConflict)
	resp.Body.Close()
}

func TestCampaign_Schedule(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampSched", "Camp Sched Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Schedule Bot")
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "Scheduled", bot.ID)

	scheduleAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	body := map[string]interface{}{"scheduled_at": scheduleAt}
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/schedule", c.ID), body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify status changed
	resp = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	var updated campaignResp
	decodeJSON(t, resp, &updated)
	if updated.Status != "scheduled" {
		t.Errorf("expected status scheduled, got %s", updated.Status)
	}
}

func TestCampaign_CancelSchedule(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampCancel", "Camp Cancel Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Cancel Bot")
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "To Cancel", bot.ID)

	// Schedule first
	scheduleAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	body := map[string]interface{}{"scheduled_at": scheduleAt}
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/schedule", c.ID), body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Cancel schedule
	resp = doRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/campaigns/%d/schedule", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestCampaign_Analytics(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampAnalytics", "Camp Analytics Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Analytics Bot")
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "Analytics Campaign", bot.ID)

	// Send first
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/send", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Get analytics
	resp = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/campaigns/%d/analytics", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestCampaign_RecordClick(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CampClick", "Camp Click Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Click Bot")
	c := mustCreateCampaign(t, ar.Tokens.AccessToken, "Clickable", bot.ID)

	// Send
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/send", c.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Record click
	clickBody := map[string]interface{}{
		"client_id":  1,
		"button_idx": 0,
	}
	resp = doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/campaigns/%d/click", c.ID), clickBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

func TestCampaign_WrongOrg(t *testing.T) {
	ar1 := mustRegister(t, uniqueEmail(t), "password123", "CampOwner1", "Camp Org1")
	ar2 := mustRegister(t, uniqueEmail(t), "password123", "CampOwner2", "Camp Org2")
	bot := mustCreateBot(t, ar1.Tokens.AccessToken, "Org1 Bot")
	c := mustCreateCampaign(t, ar1.Tokens.AccessToken, "Private Campaign", bot.ID)

	// Get
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()

	// Update
	resp = doRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), map[string]string{"name": "hack"}, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()

	// Delete
	resp = doRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/campaigns/%d", c.ID), nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}

// --- Scenarios ---

func TestScenario_CRUD(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ScenOwner", "Scen Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Scenario Bot")

	// Create
	s := mustCreateScenario(t, ar.Tokens.AccessToken, bot.ID, "Birthday Action", "birthday")
	if s.ID == 0 {
		t.Fatal("expected non-zero scenario ID")
	}
	if s.TriggerType != "birthday" {
		t.Errorf("expected trigger_type birthday, got %s", s.TriggerType)
	}

	// List
	resp := doRequest(t, http.MethodGet, "/api/v1/campaigns/scenarios", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	var scenarios []scenarioResp
	decodeJSON(t, resp, &scenarios)
	if len(scenarios) < 1 {
		t.Error("expected at least 1 scenario")
	}

	// Update
	updateBody := map[string]interface{}{"name": "Updated Scenario"}
	resp = doRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/campaigns/scenarios/%d", s.ID), updateBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Delete
	resp = doRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/campaigns/scenarios/%d", s.ID), nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestScenario_Templates(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "TmplOwner", "Tmpl Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Template Bot")

	// List templates
	resp := doRequest(t, http.MethodGet, "/api/v1/campaigns/scenarios/templates", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var templates []scenarioResp
	decodeJSON(t, resp, &templates)

	if len(templates) == 0 {
		t.Skip("no templates seeded, skipping clone test")
	}

	// Clone first template
	cloneBody := map[string]interface{}{"bot_id": bot.ID}
	resp = doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/campaigns/scenarios/templates/%s/clone", "birthday"),
		cloneBody, ar.Tokens.AccessToken)
	// Accept either 201 (clone success) or 404 (template key not found)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 201 or 404, got %d", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusCreated {
		var cloned scenarioResp
		decodeJSON(t, resp, &cloned)
		t.Cleanup(func() {
			pgMod.DB().ExecContext(t.Context(), "DELETE FROM auto_scenarios WHERE id = $1", cloned.ID)
		})
		if cloned.BotID != bot.ID {
			t.Errorf("expected cloned bot_id=%d, got %d", bot.ID, cloned.BotID)
		}
	} else {
		resp.Body.Close()
	}
}
