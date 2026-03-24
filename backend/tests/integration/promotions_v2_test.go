//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

// promotionResp and promoCodeResp are defined in promotions_test.go

type promoValidationResp struct {
	Valid    bool    `json:"valid"`
	Discount float64 `json:"discount,omitempty"`
	Message  string  `json:"message,omitempty"`
}

type generateCodeResp struct {
	Code string `json:"code"`
}

func mustCreatePromoCode(t *testing.T, token, code string, promoID int) promoCodeResp {
	t.Helper()

	body := map[string]interface{}{
		"code":         code,
		"promotion_id": promoID,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/promotions/promo-codes", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var pc promoCodeResp
	decodeJSON(t, resp, &pc)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(), "DELETE FROM promo_codes WHERE id = $1", pc.ID)
	})

	return pc
}

// --- Promo Codes v2 ---

func TestPromoCode_Generate(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "GenOwner", "Gen Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/promotions/promo-codes/generate", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var gen generateCodeResp
	decodeJSON(t, resp, &gen)

	if gen.Code == "" {
		t.Fatal("expected non-empty generated code")
	}
	if len(gen.Code) < 4 {
		t.Errorf("generated code too short: %q", gen.Code)
	}
}

func TestPromoCode_Validate_Success(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ValOwner", "Val Org")
	promo := mustCreatePromotion(t, ar.Tokens.AccessToken, "Validate Promo", "discount")
	mustCreatePromoCode(t, ar.Tokens.AccessToken, "VALID10", promo.ID)

	body := map[string]interface{}{
		"code":      "VALID10",
		"client_id": 1,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/promotions/promo-codes/validate", body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestPromoCode_Activate(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ActOwner", "Act Org")
	promo := mustCreatePromotion(t, ar.Tokens.AccessToken, "Activate Promo", "discount")
	mustCreatePromoCode(t, ar.Tokens.AccessToken, "ACTIVATE10", promo.ID)

	body := map[string]interface{}{
		"code":      "ACTIVATE10",
		"client_id": 1,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/promotions/promo-codes/activate", body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestPromoCode_Deactivate(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "DeactOwner", "Deact Org")
	promo := mustCreatePromotion(t, ar.Tokens.AccessToken, "Deact Promo", "discount")
	pc := mustCreatePromoCode(t, ar.Tokens.AccessToken, "DEACT10", promo.ID)

	resp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/promotions/promo-codes/%d", pc.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestPromoCode_ChannelAnalytics(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ChanOwner", "Chan Org")

	resp := doRequest(t, http.MethodGet, "/api/v1/promotions/promo-codes/analytics", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestPromoCode_ByPromotion(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ByPromoOwner", "ByPromo Org")
	promo := mustCreatePromotion(t, ar.Tokens.AccessToken, "Codes Promo", "discount")
	mustCreatePromoCode(t, ar.Tokens.AccessToken, "CODE1", promo.ID)
	mustCreatePromoCode(t, ar.Tokens.AccessToken, "CODE2", promo.ID)

	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/promotions/%d/codes", promo.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var codes []promoCodeResp
	decodeJSON(t, resp, &codes)
	if len(codes) < 2 {
		t.Errorf("expected at least 2 promo codes, got %d", len(codes))
	}
}
