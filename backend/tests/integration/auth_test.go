//go:build integration

package integration_test

import (
	"net/http"
	"testing"
)

func TestAuth_Register(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Test User", "Test Org")

	if ar.Tokens.AccessToken == "" {
		t.Fatal("expected access_token, got empty string")
	}
	if ar.Tokens.RefreshToken == "" {
		t.Fatal("expected refresh_token, got empty string")
	}
	if ar.User.Email != email {
		t.Fatalf("expected email %q, got %q", email, ar.User.Email)
	}
	if ar.User.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if ar.User.OrgID == 0 {
		t.Fatal("expected non-zero org ID")
	}
}

func TestAuth_Register_DuplicateEmail(t *testing.T) {
	email := uniqueEmail(t)
	mustRegister(t, email, "password123", "User One", "Org One")

	body := map[string]string{
		"email":        email,
		"password":     "password456",
		"name":         "User Two",
		"organization": "Org Two",
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/register", body, "")
	assertStatus(t, resp, http.StatusConflict)
	resp.Body.Close()
}

func TestAuth_Login(t *testing.T) {
	email := uniqueEmail(t)
	mustRegister(t, email, "password123", "Test User", "Test Org")

	ar := mustLogin(t, email, "password123")
	if ar.Tokens.AccessToken == "" {
		t.Fatal("expected access_token after login")
	}
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	email := uniqueEmail(t)
	mustRegister(t, email, "correct-password", "Test User", "Test Org")

	body := map[string]string{"email": email, "password": "wrong-password"}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/login", body, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestAuth_Login_UnknownEmail(t *testing.T) {
	body := map[string]string{"email": "nobody@test.revisitr.local", "password": "pass"}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/login", body, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestAuth_Refresh(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Test User", "Test Org")

	body := map[string]string{"refresh_token": ar.Tokens.RefreshToken}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/refresh", body, "")
	assertStatus(t, resp, http.StatusOK)

	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decodeJSON(t, resp, &tokens)

	if tokens.AccessToken == "" {
		t.Fatal("expected new access_token after refresh")
	}
	if tokens.RefreshToken == "" {
		t.Fatal("expected new refresh_token after refresh")
	}
	// Refresh token must rotate (new one issued)
	if tokens.RefreshToken == ar.Tokens.RefreshToken {
		t.Error("expected a new refresh_token, got the same one")
	}
	// New access token must be usable on protected routes
	resp2 := doRequest(t, http.MethodGet, "/api/v1/bots", nil, tokens.AccessToken)
	assertStatus(t, resp2, http.StatusOK)
	resp2.Body.Close()
}

func TestAuth_Refresh_InvalidToken(t *testing.T) {
	body := map[string]string{"refresh_token": "not-a-real-token"}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/refresh", body, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestAuth_Refresh_TokenRotation(t *testing.T) {
	// After refresh, the old refresh token must no longer work.
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Test User", "Test Org")

	// First refresh — consumes ar.Tokens.RefreshToken
	body := map[string]string{"refresh_token": ar.Tokens.RefreshToken}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/refresh", body, "")
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Second refresh with the same (now revoked) token must fail
	resp2 := doRequest(t, http.MethodPost, "/api/v1/auth/refresh", body, "")
	assertStatus(t, resp2, http.StatusUnauthorized)
	resp2.Body.Close()
}

func TestAuth_Logout(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Test User", "Test Org")

	body := map[string]string{"refresh_token": ar.Tokens.RefreshToken}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/logout", body, "")
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Refresh token must no longer work after logout
	resp2 := doRequest(t, http.MethodPost, "/api/v1/auth/refresh", body, "")
	assertStatus(t, resp2, http.StatusUnauthorized)
	resp2.Body.Close()
}

func TestAuth_ProtectedRoute_WithoutToken(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/bots", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestAuth_ProtectedRoute_WithToken(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "Test User", "Test Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/bots", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}
