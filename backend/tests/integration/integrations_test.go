//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

type integrationResp struct {
	ID     int    `json:"id"`
	OrgID  int    `json:"org_id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

func mustCreateIntegration(t *testing.T, token, intgType string) integrationResp {
	t.Helper()

	body := map[string]interface{}{
		"type":   intgType,
		"config": map[string]interface{}{},
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/integrations", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var intg integrationResp
	decodeJSON(t, resp, &intg)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(
			t.Context(),
			"DELETE FROM integrations WHERE id = $1", intg.ID,
		)
	})

	return intg
}

func TestIntegrations_RequiresAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/integrations", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestIntegrations_Create_SetsInactive(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "IntgOwner", "Intg Org")

	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "iiko")

	if intg.ID == 0 {
		t.Fatal("expected non-zero integration ID")
	}
	if intg.Status != "inactive" {
		t.Errorf("expected status=inactive, got %s", intg.Status)
	}
	if intg.OrgID != ar.User.OrgID {
		t.Errorf("expected org_id=%d, got %d", ar.User.OrgID, intg.OrgID)
	}
}

func TestIntegrations_List(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "IntgOwner2", "Intg Org2")

	mustCreateIntegration(t, ar.Tokens.AccessToken, "iiko")
	mustCreateIntegration(t, ar.Tokens.AccessToken, "rkeeper")

	resp := doRequest(t, http.MethodGet, "/api/v1/integrations", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var intgs []integrationResp
	decodeJSON(t, resp, &intgs)

	if len(intgs) < 2 {
		t.Errorf("expected at least 2 integrations, got %d", len(intgs))
	}
}

func TestIntegrations_GetByID(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "IntgOwner3", "Intg Org3")

	created := mustCreateIntegration(t, ar.Tokens.AccessToken, "iiko")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d", created.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var intg integrationResp
	decodeJSON(t, resp, &intg)

	if intg.ID != created.ID {
		t.Errorf("expected integration ID %d, got %d", created.ID, intg.ID)
	}
}

func TestIntegrations_Delete(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "IntgOwner4", "Intg Org4")

	body := map[string]interface{}{
		"type":   "iiko",
		"config": map[string]interface{}{},
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/integrations", body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var intg integrationResp
	decodeJSON(t, resp, &intg)

	delResp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/integrations/%d", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()
}

func TestIntegrations_SyncNow(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "IntgOwner5", "Intg Org5")

	intg := mustCreateIntegration(t, ar.Tokens.AccessToken, "iiko")

	syncResp := doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/integrations/%d/sync", intg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, syncResp, http.StatusOK)
	syncResp.Body.Close()
}

func TestIntegrations_CrossOrg_Forbidden(t *testing.T) {
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "IOOwner1", "IO Org1")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "IOOwner2", "IO Org2")

	intg := mustCreateIntegration(t, ar1.Tokens.AccessToken, "iiko")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/integrations/%d", intg.ID),
		nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}
