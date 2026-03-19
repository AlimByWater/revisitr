//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func doRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, srv.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, dst interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected status %d, got %d; body: %s", want, resp.StatusCode, body)
	}
}

// ── Auth helpers ──────────────────────────────────────────────────────────────

type authResp struct {
	User struct {
		ID    int    `json:"id"`
		OrgID int    `json:"org_id"`
		Email string `json:"email"`
	} `json:"user"`
	Tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	} `json:"tokens"`
}

// mustRegister registers a user and returns the auth response.
// Registers t.Cleanup to delete the user and their organization.
func mustRegister(t *testing.T, email, password, name, org string) authResp {
	t.Helper()

	body := map[string]string{
		"email":        email,
		"password":     password,
		"name":         name,
		"organization": org,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/register", body, "")
	assertStatus(t, resp, http.StatusCreated)

	var ar authResp
	decodeJSON(t, resp, &ar)

	// Cleanup: delete user and their org from the real DB
	t.Cleanup(func() {
		ctx := context.Background()
		db := pgMod.DB()
		db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", ar.User.ID)
		db.ExecContext(ctx, "DELETE FROM organizations WHERE id = $1", ar.User.OrgID)
		rdsMod.Client().Del(ctx, fmt.Sprintf("user_sessions:%d", ar.User.ID))
	})

	return ar
}

// mustLogin logs in and returns the auth response.
func mustLogin(t *testing.T, email, password string) authResp {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/login", body, "")
	assertStatus(t, resp, http.StatusOK)

	var ar authResp
	decodeJSON(t, resp, &ar)
	return ar
}

// ── Unique test data generators ───────────────────────────────────────────────

// uniqueEmail generates a unique email for a test using nanoseconds.
func uniqueEmail(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("test-%d%s", time.Now().UnixNano(), testEmailDomain)
}
