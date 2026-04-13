//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	redisRepo "revisitr/internal/repository/redis"
)

type activationLinkResp struct {
	DeepLink  string    `json:"deep_link"`
	ExpiresAt time.Time `json:"expires_at"`
}

type createManagedBotResp struct {
	BotID    int    `json:"bot_id"`
	DeepLink string `json:"deep_link"`
	Status   string `json:"status"`
}

type botStatusResp struct {
	Status string `json:"status"`
}

type botSettingsResp struct {
	Modules          []string        `json:"modules"`
	WelcomeMessage   string          `json:"welcome_message"`
	RegistrationForm []formFieldResp `json:"registration_form"`
}

type formFieldResp struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type managedBotRow struct {
	ID       int    `db:"id"`
	OrgID    int    `db:"org_id"`
	Name     string `db:"name"`
	Username string `db:"username"`
	Status   string `db:"status"`
}

func parseStartToken(t *testing.T, deepLink string) string {
	t.Helper()

	parsed, err := url.Parse(deepLink)
	if err != nil {
		t.Fatalf("parse deep link %q: %v", deepLink, err)
	}

	token := parsed.Query().Get("start")
	if token == "" {
		t.Fatalf("expected start token in %q", deepLink)
	}

	return token
}

func mustCreateManagedBot(t *testing.T, token string, body map[string]any) createManagedBotResp {
	t.Helper()

	resp := doRequest(t, http.MethodPost, "/api/v1/bots/create-managed", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var created createManagedBotResp
	decodeJSON(t, resp, &created)

	t.Cleanup(func() {
		_, _ = pgMod.DB().ExecContext(t.Context(), "DELETE FROM bots WHERE id = $1", created.BotID)
	})

	return created
}

func TestManagedBots_ActivationLink_StoresOneTimeToken(t *testing.T) {
	email := uniqueEmail(t)
	auth := mustRegister(t, email, "password123", "Managed Owner", "Managed Org")

	resp := doRequest(t, http.MethodPost, "/api/v1/bots/activation-link", nil, auth.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var body activationLinkResp
	decodeJSON(t, resp, &body)

	prefix := "https://t.me/" + testMasterBotUsername + "?start="
	if !strings.HasPrefix(body.DeepLink, prefix) {
		t.Fatalf("deep_link = %q, want prefix %q", body.DeepLink, prefix)
	}
	if diff := time.Until(body.ExpiresAt); diff < 14*time.Minute || diff > 16*time.Minute {
		t.Fatalf("expires_at delta = %s, want about 15m", diff)
	}

	token := parseStartToken(t, body.DeepLink)
	authRepo := redisRepo.NewMasterBotAuth(rdsMod)

	stored, err := authRepo.ValidateAndConsume(context.Background(), token)
	if err != nil {
		t.Fatalf("validate and consume auth token: %v", err)
	}
	if stored.OrgID != auth.User.OrgID || stored.UserID != auth.User.ID {
		t.Fatalf("stored auth token = %+v, want org_id=%d user_id=%d", stored, auth.User.OrgID, auth.User.ID)
	}

	if _, err := authRepo.ValidateAndConsume(context.Background(), token); err == nil {
		t.Fatal("expected auth token to be one-time and no longer available")
	}
}

func TestManagedBots_CreateManaged_EndToEndFlow(t *testing.T) {
	email := uniqueEmail(t)
	auth := mustRegister(t, email, "password123", "Managed Owner", "Managed Org")

	created := mustCreateManagedBot(t, auth.Tokens.AccessToken, map[string]any{
		"name":            "ManagedLaunchBot",
		"username":        "@managedlaunchbot",
		"modules":         []string{"loyalty", "campaigns"},
		"welcome_message": "Welcome to managed bot",
		"registration_form": []map[string]any{
			{"name": "phone", "label": "Phone", "type": "phone", "required": true},
		},
	})

	if created.BotID == 0 {
		t.Fatal("expected non-zero bot_id")
	}
	if created.Status != "pending" {
		t.Fatalf("status = %q, want %q", created.Status, "pending")
	}

	wantDeepLink := "https://t.me/newbot/" + testMasterBotUsername + "/managedlaunchbot?name=ManagedLaunchBot"
	if created.DeepLink != wantDeepLink {
		t.Fatalf("deep_link = %q, want %q", created.DeepLink, wantDeepLink)
	}

	var row managedBotRow
	if err := pgMod.DB().GetContext(t.Context(), &row,
		"SELECT id, org_id, name, username, status FROM bots WHERE id = $1",
		created.BotID,
	); err != nil {
		t.Fatalf("fetch managed bot row: %v", err)
	}
	if row.OrgID != auth.User.OrgID {
		t.Fatalf("org_id = %d, want %d", row.OrgID, auth.User.OrgID)
	}
	if row.Username != "managedlaunchbot" {
		t.Fatalf("username = %q, want %q", row.Username, "managedlaunchbot")
	}
	if row.Status != "pending" {
		t.Fatalf("row status = %q, want %q", row.Status, "pending")
	}

	statusResp := doRequest(
		t,
		http.MethodGet,
		"/api/v1/bots/"+strconv.Itoa(created.BotID)+"/status",
		nil,
		auth.Tokens.AccessToken,
	)
	assertStatus(t, statusResp, http.StatusOK)

	var statusBody botStatusResp
	decodeJSON(t, statusResp, &statusBody)
	if statusBody.Status != "pending" {
		t.Fatalf("status endpoint = %q, want %q", statusBody.Status, "pending")
	}

	settingsResp := doRequest(
		t,
		http.MethodGet,
		"/api/v1/bots/"+strconv.Itoa(created.BotID)+"/settings",
		nil,
		auth.Tokens.AccessToken,
	)
	assertStatus(t, settingsResp, http.StatusOK)

	var settingsBody botSettingsResp
	decodeJSON(t, settingsResp, &settingsBody)
	if len(settingsBody.Modules) != 2 {
		t.Fatalf("modules len = %d, want 2", len(settingsBody.Modules))
	}
	if settingsBody.WelcomeMessage != "Welcome to managed bot" {
		t.Fatalf("welcome_message = %q, want %q", settingsBody.WelcomeMessage, "Welcome to managed bot")
	}
	if len(settingsBody.RegistrationForm) != 1 || settingsBody.RegistrationForm[0].Name != "phone" {
		t.Fatalf("registration_form = %#v, want single phone field", settingsBody.RegistrationForm)
	}
}

func TestManagedBots_CreateManaged_RejectsInvalidUsername(t *testing.T) {
	email := uniqueEmail(t)
	auth := mustRegister(t, email, "password123", "Managed Owner", "Managed Org")

	resp := doRequest(t, http.MethodPost, "/api/v1/bots/create-managed", map[string]any{
		"name":     "ManagedLaunchBot",
		"username": "@managedlaunch",
	}, auth.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func TestManagedBots_Status_ForbiddenForAnotherOrg(t *testing.T) {
	email1 := uniqueEmail(t)
	auth1 := mustRegister(t, email1, "password123", "Managed Owner", "Managed Org")
	created := mustCreateManagedBot(t, auth1.Tokens.AccessToken, map[string]any{
		"name":     "ManagedLaunchBot",
		"username": "@managedstatusbot",
	})

	email2 := uniqueEmail(t)
	auth2 := mustRegister(t, email2, "password123", "Other Owner", "Other Org")

	resp := doRequest(
		t,
		http.MethodGet,
		"/api/v1/bots/"+strconv.Itoa(created.BotID)+"/status",
		nil,
		auth2.Tokens.AccessToken,
	)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}
