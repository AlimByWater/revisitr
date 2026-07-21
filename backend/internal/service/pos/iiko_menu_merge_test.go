package pos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"revisitr/internal/entity"
)

func TestIikoGetMenuUsesExternalMenuWhenConfigured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/1/access_token":
			_ = json.NewEncoder(w).Encode(iikoTokenResponse{Token: "tok"})
		case "/api/2/menu/by_id":
			price := 250.0
			_ = json.NewEncoder(w).Encode(iikoExternalMenuResponse{
				ItemCategories: []iikoExternalMenuCategory{{
					Name: "Coffee",
					Items: []iikoExternalMenuItem{{
						ItemID:    "product-1",
						Name:      "Cappuccino",
						ItemSizes: []iikoExternalMenuItemSize{{Price: &price}},
					}},
				}},
			})
		case "/api/1/nomenclature":
			t.Fatal("nomenclature must not be requested when external_menu_id is configured")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	provider, err := NewIikoProvider(
		entity.IntegrationConfig{APIKey: "key", OrgID: "org-1", ExternalMenuID: "menu-1"},
		WithIikoBaseURL(srv.URL+"/api/1"),
		WithIikoHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("NewIikoProvider: %v", err)
	}

	menu, err := provider.GetMenu(context.Background())
	if err != nil {
		t.Fatalf("GetMenu: %v", err)
	}
	if len(menu.Categories) != 1 || len(menu.Categories[0].Items) != 1 {
		t.Fatalf("menu=%+v", menu)
	}
	if item := menu.Categories[0].Items[0]; item.ExternalID != "product-1" || item.Name != "Cappuccino" || item.Price != 250 {
		t.Fatalf("item=%+v", item)
	}
}
