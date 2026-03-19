//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

type promotionResp struct {
	ID     int    `json:"id"`
	OrgID  int    `json:"org_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Active bool   `json:"active"`
}

type promoCodeResp struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
}

func mustCreatePromotion(t *testing.T, token, name, pType string) promotionResp {
	t.Helper()

	body := map[string]interface{}{
		"name": name,
		"type": pType,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/promotions", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var p promotionResp
	decodeJSON(t, resp, &p)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(
			t.Context(),
			"DELETE FROM promotions WHERE id = $1", p.ID,
		)
	})

	return p
}

func TestPromotions_RequiresAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/api/v1/promotions", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestPromotions_Create_SetsActive(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "PromoOwner", "Promo Org")

	p := mustCreatePromotion(t, ar.Tokens.AccessToken, "Summer Deal", "discount")

	if p.ID == 0 {
		t.Fatal("expected non-zero promotion ID")
	}
	if !p.Active {
		t.Error("expected promotion to be active on creation")
	}
	if p.OrgID != ar.User.OrgID {
		t.Errorf("expected org_id=%d, got %d", ar.User.OrgID, p.OrgID)
	}
}

func TestPromotions_List(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "PromoOwner2", "Promo Org2")

	mustCreatePromotion(t, ar.Tokens.AccessToken, "Promo A", "discount")
	mustCreatePromotion(t, ar.Tokens.AccessToken, "Promo B", "bonus")

	resp := doRequest(t, http.MethodGet, "/api/v1/promotions", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var promos []promotionResp
	decodeJSON(t, resp, &promos)

	if len(promos) < 2 {
		t.Errorf("expected at least 2 promotions, got %d", len(promos))
	}
}

func TestPromotions_GetByID(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "PromoOwner3", "Promo Org3")

	created := mustCreatePromotion(t, ar.Tokens.AccessToken, "My Promo", "discount")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/promotions/%d", created.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var p promotionResp
	decodeJSON(t, resp, &p)

	if p.ID != created.ID {
		t.Errorf("expected promotion ID %d, got %d", created.ID, p.ID)
	}
}

func TestPromotions_Delete(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "PromoOwner4", "Promo Org4")

	body := map[string]interface{}{
		"name": "To Delete",
		"type": "discount",
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/promotions", body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var p promotionResp
	decodeJSON(t, resp, &p)

	delResp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/promotions/%d", p.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()
}

func TestPromotions_PromoCode_Create(t *testing.T) {
	email := uniqueEmail(t)
	ar := mustRegister(t, email, "password123", "PromoOwner5", "Promo Org5")

	promo := mustCreatePromotion(t, ar.Tokens.AccessToken, "Code Promo", "discount")

	body := map[string]interface{}{
		"code":            "SAVE10",
		"promotion_id":    promo.ID,
	}
	resp := doRequest(t, http.MethodPost,
		"/api/v1/promotions/promo-codes",
		body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var pc promoCodeResp
	decodeJSON(t, resp, &pc)

	if pc.ID == 0 {
		t.Fatal("expected non-zero promo code ID")
	}

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(), "DELETE FROM promo_codes WHERE id = $1", pc.ID)
	})
}

func TestPromotions_CrossOrg_Forbidden(t *testing.T) {
	email1 := uniqueEmail(t)
	ar1 := mustRegister(t, email1, "password123", "POOwner1", "PO Org1")

	email2 := uniqueEmail(t)
	ar2 := mustRegister(t, email2, "password123", "POOwner2", "PO Org2")

	p := mustCreatePromotion(t, ar1.Tokens.AccessToken, "Private Promo", "discount")

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/promotions/%d", p.ID),
		nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}
