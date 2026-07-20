package pos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"revisitr/internal/entity"
)

func TestIikoGetMenuUsesNomenclatureAndEnrichesFromExternalMenu(t *testing.T) {
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
					Description: "Original description",
					ParentGroup: "group-1",
					Price:       &price,
				}, {
					ID:          "product-3",
					Name:        "Americano",
					ParentGroup: "group-1",
					Price:       &price,
				}},
			})
		case "/api/2/menu/by_id":
			price := 250.0
			_ = json.NewEncoder(w).Encode(iikoExternalMenuResponse{
				ItemCategories: []iikoExternalMenuCategory{{
					Name: "Coffee",
					Items: []iikoExternalMenuItem{{
						IikoItemID:  "product-1",
						Name:        "External Cappuccino",
						Description: "External description",
						ImageLinks:  []string{"https://iiko.example/cappuccino.jpg"},
						ItemSizes:   []iikoExternalMenuItemSize{{Price: &price}},
					}},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	provider, err := NewIikoProvider(
		entity.IntegrationConfig{APIKey: "test-login", OrgID: "org-1", ExternalMenuID: "menu-1"},
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
	if len(menu.Categories) != 1 || len(menu.Categories[0].Items) != 2 {
		t.Fatalf("menu=%+v", menu)
	}

	items := menu.Categories[0].Items
	if items[0].ExternalID != "product-3" || items[0].Name != "Americano" || items[0].Price != 190 {
		t.Fatalf("nomenclature-only item=%+v", items[0])
	}
	if items[1].ExternalID != "product-1" || items[1].Name != "iiko Cappuccino" || items[1].Description != "External description" || items[1].ImageURL != "https://iiko.example/cappuccino.jpg" || items[1].Price != 250 {
		t.Fatalf("enriched item=%+v", items[1])
	}
}
