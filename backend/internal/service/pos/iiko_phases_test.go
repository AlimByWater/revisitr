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

// deliveryDay builds a one-order delivery response for a given day/sum/phone.
func deliveryDay(id, phone string, sum float64, whenClosed string) iikoDeliveryOrder {
	return iikoDeliveryOrder{
		ID: id,
		Order: &iikoDeliveryBody{
			Phone:      phone,
			Status:     "Closed",
			Sum:        sum,
			WhenClosed: whenClosed,
			Items: []iikoDeliveryItem{{
				Type:      "Product",
				Amount:    1,
				Price:     sum,
				Product:   &iikoDeliveryProduct{ID: "p-1", Name: "Coffee"},
				ResultSum: sum,
			}},
		},
	}
}

// TestIiko_GetOrders_ChunksWindow verifies that a multi-day window is split into
// per-day requests (avoiding iiko 422 TOO_MANY_DATA_REQUESTED) and merged.
func TestIiko_GetOrders_ChunksWindow(t *testing.T) {
	t.Parallel()

	var deliveryCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/deliveries/by_delivery_date_and_status":
			n := atomic.AddInt32(&deliveryCalls, 1)
			body, _ := io.ReadAll(r.Body)
			// Each chunk must be <= 24h. Reject wider windows like the real API.
			var req struct {
				From string `json:"deliveryDateFrom"`
				To   string `json:"deliveryDateTo"`
			}
			_ = json.Unmarshal(body, &req)
			from, _ := time.Parse("2006-01-02 15:04:05.000", req.From)
			to, _ := time.Parse("2006-01-02 15:04:05.000", req.To)
			if to.Sub(from) > 24*time.Hour+time.Second {
				http.Error(w, `{"error":"TOO_MANY_DATA_REQUESTED"}`, http.StatusUnprocessableEntity)
				return
			}
			// Return one order on the first chunk only.
			resp := iikoDeliveriesResponse{}
			if n == 1 {
				resp.OrdersByOrganizations = []iikoOrdersByOrganization{{
					OrganizationID: "org-1",
					Orders:         []iikoDeliveryOrder{deliveryDay("order-1", "+79990000000", 500, "2026-05-28 12:30:00.000")},
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
	to := from.Add(3 * 24 * time.Hour) // 3-day window, padded ±24h → 5 daily chunks

	orders, err := p.GetOrders(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if got := atomic.LoadInt32(&deliveryCalls); got != 5 {
		t.Fatalf("expected 5 chunked delivery calls (3-day window + ±24h timezone pad), got %d", got)
	}
	if len(orders) != 1 || orders[0].ExternalID != "order-1" {
		t.Fatalf("orders=%+v", orders)
	}
}

// TestIiko_GetOrders_PadsWindowForTimezone verifies the deliveries window is
// widened ±24h before being sent to iiko. iiko filters by each order's delivery
// date in the org's local timezone while we pass UTC, so without the pad orders
// from the last few hours fall past the bounds and are skipped. See
// iikoOrderWindowPad.
func TestIiko_GetOrders_PadsWindowForTimezone(t *testing.T) {
	t.Parallel()

	// fetchDeliveriesChunked requests chunks sequentially, so no locking needed.
	var minFrom, maxTo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/deliveries/by_delivery_date_and_status":
			var req struct {
				From string `json:"deliveryDateFrom"`
				To   string `json:"deliveryDateTo"`
			}
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &req)
			if minFrom == "" || req.From < minFrom {
				minFrom = req.From
			}
			if req.To > maxTo {
				maxTo = req.To
			}
			_ = json.NewEncoder(w).Encode(iikoDeliveriesResponse{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	if _, err := p.GetOrders(context.Background(), from, to); err != nil {
		t.Fatalf("GetOrders: %v", err)
	}

	if want := iikoDateTime(from.Add(-iikoOrderWindowPad)); minFrom != want {
		t.Errorf("earliest deliveryDateFrom = %q, want %q (from - pad)", minFrom, want)
	}
	if want := iikoDateTime(to.Add(iikoOrderWindowPad)); maxTo != want {
		t.Errorf("latest deliveryDateTo = %q, want %q (to + pad)", maxTo, want)
	}
}

// TestIiko_GetDailyAggregates groups orders per day with revenue/avg/tx/guests.
func TestIiko_GetDailyAggregates(t *testing.T) {
	t.Parallel()

	// Map each per-day chunk (by from-date) to its orders.
	dayOrders := map[string][]iikoDeliveryOrder{
		"2026-05-28": {
			deliveryDay("o-1", "+79990000001", 300, "2026-05-28 10:00:00.000"),
			deliveryDay("o-2", "+79990000002", 500, "2026-05-28 20:00:00.000"),
		},
		"2026-05-29": {
			deliveryDay("o-3", "+79990000001", 200, "2026-05-29 11:00:00.000"),
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/deliveries/by_delivery_date_and_status":
			body, _ := io.ReadAll(r.Body)
			var req struct {
				From string `json:"deliveryDateFrom"`
			}
			_ = json.Unmarshal(body, &req)
			day := req.From[:10]
			_ = json.NewEncoder(w).Encode(iikoDeliveriesResponse{
				OrdersByOrganizations: []iikoOrdersByOrganization{{
					OrganizationID: "org-1",
					Orders:         dayOrders[day],
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	to := from.Add(2 * 24 * time.Hour)

	aggs, err := p.GetDailyAggregates(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetDailyAggregates: %v", err)
	}
	if len(aggs) != 2 {
		t.Fatalf("expected 2 days, got %d: %+v", len(aggs), aggs)
	}

	day1 := aggs[0]
	if day1.Date.Format("2006-01-02") != "2026-05-28" {
		t.Fatalf("day1 date=%s", day1.Date)
	}
	if day1.Revenue != 800 || day1.TxCount != 2 || day1.AvgCheck != 400 || day1.GuestCount != 2 {
		t.Fatalf("day1=%+v", day1)
	}

	day2 := aggs[1]
	if day2.Revenue != 200 || day2.TxCount != 1 || day2.AvgCheck != 200 || day2.GuestCount != 1 {
		t.Fatalf("day2=%+v", day2)
	}
}

// TestIiko_GetCustomers_ByPhone resolves a loyalty customer by phone.
func TestIiko_GetCustomers_ByPhone(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/loyalty/iiko/customer/info":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"phone":"+79990000001"`) {
				t.Errorf("unexpected body: %s", body)
			}
			_ = json.NewEncoder(w).Encode(iikoCustomerInfo{
				ID:             "cust-1",
				Name:           "Иван",
				Surname:        "Петров",
				Phone:          "+79990000001",
				Email:          "ivan@example.com",
				Birthday:       "1990-05-15 00:00:00.000",
				Cards:          []iikoCustomerCard{{Number: "CARD-1"}},
				WalletBalances: []iikoCustomerWallet{{Balance: 120}, {Balance: 80}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	customers, err := p.GetCustomers(context.Background(), CustomerListOpts{Search: "+79990000001"})
	if err != nil {
		t.Fatalf("GetCustomers: %v", err)
	}
	if len(customers) != 1 {
		t.Fatalf("expected 1 customer, got %d", len(customers))
	}
	c := customers[0]
	if c.ExternalID != "cust-1" || c.Name != "Иван Петров" || c.Phone != "+79990000001" {
		t.Fatalf("customer=%+v", c)
	}
	if c.Balance != 200 || c.CardNumber != "CARD-1" || c.Email != "ivan@example.com" {
		t.Fatalf("customer fields=%+v", c)
	}
	if c.Birthday == nil || c.Birthday.Format("2006-01-02") != "1990-05-15" {
		t.Fatalf("birthday=%v", c.Birthday)
	}
}

// TestIiko_GetCustomers_NoSearch returns nil when no phone is provided (no bulk list).
func TestIiko_GetCustomers_NoSearch(t *testing.T) {
	t.Parallel()

	p := newIikoTestProvider(t, httptest.NewServer(http.NotFoundHandler()))
	customers, err := p.GetCustomers(context.Background(), CustomerListOpts{})
	if err != nil {
		t.Fatalf("GetCustomers: %v", err)
	}
	if customers != nil {
		t.Fatalf("expected nil, got %+v", customers)
	}
}

// TestIiko_GetOrders_SkipsDegenerateTrailingWindow reproduces the prod bug where
// a 30-day window produced a sub-millisecond trailing chunk whose from/to render
// to the same instant → iiko 500 "deliveryDateFrom must be less than deliveryDateTo".
// The chunker must skip such windows instead of requesting them.
func TestIiko_GetOrders_SkipsDegenerateTrailingWindow(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/deliveries/by_delivery_date_and_status":
			body, _ := io.ReadAll(r.Body)
			var req struct {
				From string `json:"deliveryDateFrom"`
				To   string `json:"deliveryDateTo"`
			}
			_ = json.Unmarshal(body, &req)
			if req.From == req.To {
				// This is exactly what iiko rejects; the chunker must never send it.
				t.Errorf("degenerate window requested: from==to==%q", req.From)
				http.Error(w, `{"errorDescription":"deliveryDateFrom must be less than deliveryDateTo"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(iikoDeliveriesResponse{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newIikoTestProvider(t, srv)
	// 30 days minus a few hundred microseconds — mimics from=now-30d, to=now
	// computed on adjacent lines, so the final chunk is sub-millisecond.
	from := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	to := from.Add(30 * 24 * time.Hour).Add(400 * time.Microsecond)

	if _, err := p.GetOrders(context.Background(), from, to); err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
}

// TestIiko_Discovery_ListsOrgsAndMenus verifies onboarding discovery: list
// organizations and external menus from credentials alone (no org_id yet).
func TestIiko_Discovery_ListsOrgsAndMenus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/1/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/api/1/organizations":
			_ = json.NewEncoder(w).Encode(iikoOrganizationsResponse{
				Organizations: []iikoOrganization{
					{ID: "org-1", Name: "Мой ресторан"},
					{ID: "org-deleted", Name: "Old", IsDeleted: true},
				},
			})
		case "/api/2/menu":
			_ = json.NewEncoder(w).Encode(iikoExternalMenuListResponse{
				ExternalMenus: []iikoExternalMenuListItem{{ID: "82279", Name: "Revisitr Demo Menu"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// org_id is intentionally a placeholder — discovery is how the user finds it.
	p, err := NewIikoProvider(
		entity.IntegrationConfig{APIKey: "test-login", OrgID: "discovery"},
		WithIikoBaseURL(srv.URL+"/api/1"),
		WithIikoHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("NewIikoProvider: %v", err)
	}

	orgs, err := p.ListOrganizations(context.Background())
	if err != nil {
		t.Fatalf("ListOrganizations: %v", err)
	}
	if len(orgs) != 1 || orgs[0].ID != "org-1" || orgs[0].Name != "Мой ресторан" {
		t.Fatalf("orgs=%+v (deleted org must be filtered)", orgs)
	}

	menus, err := p.ListExternalMenus(context.Background())
	if err != nil {
		t.Fatalf("ListExternalMenus: %v", err)
	}
	if len(menus) != 1 || menus[0].ID != "82279" || menus[0].Name != "Revisitr Demo Menu" {
		t.Fatalf("menus=%+v", menus)
	}
}

// TestIiko_ListExternalMenus_GracefulWhenForbidden returns nil (not error) when
// menu listing is not permitted, so discovery still yields organizations.
func TestIiko_ListExternalMenus_GracefulWhenForbidden(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/1/access_token" {
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
			return
		}
		http.Error(w, `{"errorDescription":"not allowed"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	p, err := NewIikoProvider(
		entity.IntegrationConfig{APIKey: "k", OrgID: "discovery"},
		WithIikoBaseURL(srv.URL+"/api/1"),
		WithIikoHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("NewIikoProvider: %v", err)
	}

	menus, err := p.ListExternalMenus(context.Background())
	if err != nil {
		t.Fatalf("expected graceful nil, got error %v", err)
	}
	if menus != nil {
		t.Fatalf("expected nil menus, got %+v", menus)
	}
}

// (iiko answers 400 WrongCrmId / 401 / 403 / 404) instead of failing sync.
func TestIiko_GetCustomers_LoyaltyUnavailable(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/access_token" {
				_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
				return
			}
			http.Error(w, `{"errorCode":"Common_OrganizationNotFound"}`, status)
		}))

		p := newIikoTestProvider(t, srv)
		customers, err := p.GetCustomers(context.Background(), CustomerListOpts{Search: "+79990000001"})
		srv.Close()
		if err != nil {
			t.Fatalf("status %d: expected graceful nil, got error %v", status, err)
		}
		if customers != nil {
			t.Fatalf("status %d: expected nil customers, got %+v", status, customers)
		}
	}
}
