package pos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"revisitr/internal/entity"
)

func newIikoTestProvider(t *testing.T, srv *httptest.Server) *IikoProvider {
	t.Helper()
	p, err := NewIikoProvider(
		entity.IntegrationConfig{APIKey: "test-login", OrgID: "org-1"},
		WithIikoBaseURL(srv.URL),
		WithIikoHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("NewIikoProvider: %v", err)
	}
	return p
}

func TestNewIikoProvider_Validation(t *testing.T) {
	t.Parallel()

	if _, err := NewIikoProvider(entity.IntegrationConfig{OrgID: "x"}); err == nil {
		t.Fatal("expected error for missing apiLogin")
	}
	if _, err := NewIikoProvider(entity.IntegrationConfig{APIKey: "x"}); err == nil {
		t.Fatal("expected error for missing orgID")
	}
	p, err := NewIikoProvider(entity.IntegrationConfig{APIKey: "k", OrgID: "o"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if p.baseURL != iikoDefaultBaseURL {
		t.Fatalf("default baseURL not applied: %s", p.baseURL)
	}
}

func TestIiko_GetToken_CachesAndRefreshes(t *testing.T) {
	t.Parallel()

	var tokenCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			atomic.AddInt32(&tokenCalls, 1)
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"apiLogin":"test-login"`) {
				t.Errorf("unexpected token body: %s", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok-1", CorrelationID: "cid"})
		default:
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)

	for i := 0; i < 3; i++ {
		tok, err := p.getToken(context.Background())
		if err != nil {
			t.Fatalf("getToken: %v", err)
		}
		if tok != "tok-1" {
			t.Fatalf("token=%q", tok)
		}
	}
	if got := atomic.LoadInt32(&tokenCalls); got != 1 {
		t.Fatalf("expected 1 token fetch, got %d", got)
	}

	p.invalidateToken()
	if _, err := p.getToken(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got := atomic.LoadInt32(&tokenCalls); got != 2 {
		t.Fatalf("expected 2 token fetches after invalidate, got %d", got)
	}
}

func TestIiko_DoRequest_Success(t *testing.T) {
	t.Parallel()

	type echoReq struct {
		Foo string `json:"foo"`
	}
	type echoResp struct {
		Echo string `json:"echo"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok-1"})
		case "/echo":
			if got := r.Header.Get("Authorization"); got != "Bearer tok-1" {
				t.Errorf("auth header=%q", got)
			}
			var req echoReq
			_ = json.NewDecoder(r.Body).Decode(&req)
			_ = json.NewEncoder(w).Encode(echoResp{Echo: req.Foo})
		default:
			http.Error(w, "nope", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	var out echoResp
	if err := p.doRequest(context.Background(), "/echo", echoReq{Foo: "bar"}, &out); err != nil {
		t.Fatalf("doRequest: %v", err)
	}
	if out.Echo != "bar" {
		t.Fatalf("echo=%q", out.Echo)
	}
}

func TestIiko_DoRequest_RetriesOn401(t *testing.T) {
	t.Parallel()

	var tokenCalls, echoCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			n := atomic.AddInt32(&tokenCalls, 1)
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok-" + itoa(n)})
		case "/echo":
			n := atomic.AddInt32(&echoCalls, 1)
			if n == 1 {
				http.Error(w, "stale token", http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer tok-2" {
				t.Errorf("retry auth header=%q", got)
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	var out map[string]any
	if err := p.doRequest(context.Background(), "/echo", nil, &out); err != nil {
		t.Fatalf("doRequest: %v", err)
	}
	if atomic.LoadInt32(&tokenCalls) != 2 {
		t.Fatalf("expected 2 token fetches, got %d", tokenCalls)
	}
	if atomic.LoadInt32(&echoCalls) != 2 {
		t.Fatalf("expected 2 echo calls, got %d", echoCalls)
	}
}

func TestIiko_DoRequest_PersistentError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/access_token" {
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
			return
		}
		http.Error(w, `{"errorDescription":"bad"}`, http.StatusBadRequest)
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	err := p.doRequest(context.Background(), "/x", map[string]any{}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*IikoAPIError)
	if !ok {
		t.Fatalf("err type=%T", err)
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Fatalf("status=%d", apiErr.Status)
	}
	if !strings.Contains(apiErr.Body, "bad") {
		t.Fatalf("body=%q", apiErr.Body)
	}
}

func TestIiko_TokenFetch_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errorDescription":"bad apiLogin"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	_, err := p.getToken(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthError(err) {
		t.Fatalf("expected auth error, got %T %v", err, err)
	}
}

func TestIiko_TestConnection_VerifiesOrganization(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/organizations":
			_ = json.NewEncoder(w).Encode(iikoOrganizationsResponse{
				Organizations: []iikoOrganization{{ID: "org-1", Name: "Demo"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	if err := p.TestConnection(context.Background()); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
}

func TestIiko_TestConnection_MissingOrganization(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/organizations":
			_ = json.NewEncoder(w).Encode(iikoOrganizationsResponse{
				Organizations: []iikoOrganization{{ID: "other", Name: "Other"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	err := p.TestConnection(context.Background())
	if err == nil || !strings.Contains(err.Error(), "organization org-1 not found") {
		t.Fatalf("expected missing org error, got %v", err)
	}
}

func TestIiko_GetMenu_ExternalMenu(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/1/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/api/1/nomenclature":
			price := 190.0
			_ = json.NewEncoder(w).Encode(iikoNomenclatureResponse{
				Groups: []iikoNomenclatureGroup{{ID: "group-1", Name: "Coffee"}},
				Products: []iikoNomenclatureProduct{{
					ID:          "product-1",
					Name:        "iiko Cappuccino",
					ParentGroup: "group-1",
					Price:       &price,
				}},
			})
		case "/api/2/menu/by_id":
			if got := r.Header.Get("Authorization"); got != "Bearer tok" {
				t.Errorf("auth header=%q", got)
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"externalMenuId":"menu-1"`) {
				t.Errorf("unexpected body: %s", body)
			}
			price := 250.0
			_ = json.NewEncoder(w).Encode(iikoExternalMenuResponse{
				ItemCategories: []iikoExternalMenuCategory{{
					Name: "Coffee",
					Items: []iikoExternalMenuItem{{
						ID:          "menu-item-1",
						IikoItemID:  "product-1",
						Name:        "Cappuccino",
						Description: "Milk coffee",
						ItemSizes:   []iikoExternalMenuItemSize{{Price: &price}},
					}, {
						ItemID:    "menu-product-2",
						Name:      "Cookie",
						ItemSizes: []iikoExternalMenuItemSize{{Prices: []iikoExternalMenuPrice{{Price: &price}}}},
					}},
				}, {
					Name:     "Hidden",
					IsHidden: true,
					Items:    []iikoExternalMenuItem{{ID: "hidden", Name: "Hidden"}},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p, err := NewIikoProvider(
		entity.IntegrationConfig{APIKey: "test-login", OrgID: "org-1", ExternalMenuID: "menu-1"},
		WithIikoBaseURL(srv.URL+"/api/1"),
		WithIikoHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("NewIikoProvider: %v", err)
	}

	menu, err := p.GetMenu(context.Background())
	if err != nil {
		t.Fatalf("GetMenu: %v", err)
	}
	if len(menu.Categories) != 1 || menu.Categories[0].Name != "Coffee" || len(menu.Categories[0].Items) != 1 {
		t.Fatalf("categories=%+v", menu.Categories)
	}
	item := menu.Categories[0].Items[0]
	if item.ExternalID != "product-1" || item.Name != "iiko Cappuccino" || item.Description != "Milk coffee" || item.Price != 250 {
		t.Fatalf("item=%+v", item)
	}
}

func TestIiko_GetMenu_Nomenclature(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/nomenclature":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"organizationId":"org-1"`) {
				t.Errorf("unexpected body: %s", body)
			}
			price := 190.0
			_ = json.NewEncoder(w).Encode(iikoNomenclatureResponse{
				Groups: []iikoNomenclatureGroup{{ID: "group-1", Name: "Coffee"}},
				Products: []iikoNomenclatureProduct{{
					ID:          "product-1",
					Name:        "Espresso",
					Description: "Shot",
					ParentGroup: "group-1",
					Price:       &price,
				}, {
					ID:        "deleted",
					Name:      "Deleted",
					IsDeleted: true,
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	menu, err := p.GetMenu(context.Background())
	if err != nil {
		t.Fatalf("GetMenu: %v", err)
	}
	if len(menu.Categories) != 1 || menu.Categories[0].Name != "Coffee" {
		t.Fatalf("categories=%+v", menu.Categories)
	}
	item := menu.Categories[0].Items[0]
	if item.ExternalID != "product-1" || item.Name != "Espresso" || item.Price != 190 {
		t.Fatalf("item=%+v", item)
	}
}

func TestIiko_GetOrders_Deliveries(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/deliveries/by_delivery_date_and_status":
			body, _ := io.ReadAll(r.Body)
			reqBody := string(body)
			if !strings.Contains(reqBody, `"organizationIds":["org-1"]`) {
				t.Errorf("unexpected org body: %s", reqBody)
			}
			if !strings.Contains(reqBody, `"statuses":["Closed","Delivered"]`) {
				t.Errorf("unexpected statuses body: %s", reqBody)
			}
			// GetOrders pads the window ±24h for timezone safety, so the range is
			// split into several daily chunks. Return the order only on the chunk
			// whose window contains its delivery date, mirroring the real API.
			var req struct {
				From string `json:"deliveryDateFrom"`
			}
			_ = json.Unmarshal(body, &req)
			resp := iikoDeliveriesResponse{}
			if strings.HasPrefix(req.From, "2026-05-28") {
				resp.OrdersByOrganizations = []iikoOrdersByOrganization{{
					OrganizationID: "org-1",
					Orders: []iikoDeliveryOrder{{
						ID: "order-1",
						Order: &iikoDeliveryBody{
							Phone:      "+79990000000",
							Status:     "Closed",
							Sum:        500,
							WhenClosed: "2026-05-28 12:30:00.000",
							Discounts:  []iikoDiscountItem{{Sum: 50}},
							Items: []iikoDeliveryItem{{
								Type:      "Product",
								Amount:    2,
								Price:     250,
								Product:   &iikoDeliveryProduct{ID: "product-1", Name: "Espresso"},
								ResultSum: 500,
							}, {
								Type:    "Product",
								Deleted: map[string]any{"reason": "removed"},
								Amount:  1,
								Product: &iikoDeliveryProduct{ID: "deleted", Name: "Deleted"},
							}},
						},
					}},
				}}
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	orders, err := p.GetOrders(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("orders=%+v", orders)
	}
	order := orders[0]
	if order.ExternalID != "order-1" || order.CustomerPhone != "+79990000000" || order.Total != 500 || order.Discount != 50 {
		t.Fatalf("order=%+v", order)
	}
	if order.OrderedAt.Format("2006-01-02 15:04:05") != "2026-05-28 12:30:00" {
		t.Fatalf("ordered_at=%s", order.OrderedAt)
	}
	if len(order.Items) != 1 || order.Items[0].ExternalID != "product-1" ||
		order.Items[0].Name != "Espresso" || order.Items[0].Quantity != 2 || order.Items[0].Price != 250 {
		t.Fatalf("items=%+v", order.Items)
	}
}

func itoa(n int32) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return "n"
}
