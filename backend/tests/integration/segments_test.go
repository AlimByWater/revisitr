//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

type segmentResp struct {
	ID         int    `json:"id"`
	OrgID      int    `json:"org_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	AutoAssign bool   `json:"auto_assign"`
}

func mustCreateSegment(t *testing.T, token, name, segType string) segmentResp {
	t.Helper()

	body := map[string]interface{}{
		"name":        name,
		"type":        segType,
		"filter":      map[string]interface{}{},
		"auto_assign": false,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/segments", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var seg segmentResp
	decodeJSON(t, resp, &seg)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(
			t.Context(),
			"DELETE FROM segments WHERE id = $1", seg.ID,
		)
	})

	return seg
}

func TestSegments_RequiresAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/segments", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestSegments_Create(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "SegOwner", "Seg Org")

	seg := mustCreateSegment(t, ar.Tokens.AccessToken, "VIP Clients", "custom")

	if seg.ID == 0 {
		t.Fatal("expected non-zero segment ID")
	}
	if seg.OrgID != ar.User.OrgID {
		t.Errorf("expected org_id=%d, got %d", ar.User.OrgID, seg.OrgID)
	}
	if seg.Name != "VIP Clients" {
		t.Errorf("expected name VIP Clients, got %s", seg.Name)
	}
}

func TestSegments_List(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "SegOwner2", "Seg Org2")

	mustCreateSegment(t, ar.Tokens.AccessToken, "Segment A", "custom")
	mustCreateSegment(t, ar.Tokens.AccessToken, "Segment B", "custom")

	resp := doRequest(t, http.MethodGet, "/api/v1/segments", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var segs []segmentResp
	decodeJSON(t, resp, &segs)

	if len(segs) < 2 {
		t.Errorf("expected at least 2 segments, got %d", len(segs))
	}
}

func TestSegments_GetByID(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "SegOwner3", "Seg Org3")

	created := mustCreateSegment(t, ar.Tokens.AccessToken, "My Segment", "custom")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/segments/%d", created.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var seg segmentResp
	decodeJSON(t, resp, &seg)

	if seg.ID != created.ID {
		t.Errorf("expected segment ID %d, got %d", created.ID, seg.ID)
	}
}

func TestSegments_GetClients(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "SegOwner4", "Seg Org4")

	seg := mustCreateSegment(t, ar.Tokens.AccessToken, "Empty Segment", "custom")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/segments/%d/clients", seg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var page struct {
		Data  []interface{} `json:"data"`
		Total int           `json:"total"`
	}
	decodeJSON(t, resp, &page)
	// Empty segment — total should be 0
}

func TestSegments_Delete(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "SegOwner5", "Seg Org5")

	body := map[string]interface{}{
		"name":        "To Delete",
		"type":        "custom",
		"filter":      map[string]interface{}{},
		"auto_assign": false,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/segments", body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var seg segmentResp
	decodeJSON(t, resp, &seg)

	delResp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/segments/%d", seg.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()
}

func TestSegments_CrossOrg_Forbidden(t *testing.T) {
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "Owner1", "Org1")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "Owner2", "Org2")

	seg := mustCreateSegment(t, ar1.Tokens.AccessToken, "Private Segment", "custom")

	// org2 tries to access org1's segment
	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/segments/%d", seg.ID),
		nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}
