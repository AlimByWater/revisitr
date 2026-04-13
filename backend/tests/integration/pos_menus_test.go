//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

type posResp struct {
	ID       int            `json:"id"`
	OrgID    int            `json:"org_id"`
	Name     string         `json:"name"`
	Address  string         `json:"address"`
	Phone    string         `json:"phone"`
	Schedule map[string]any `json:"schedule"`
	IsActive bool           `json:"is_active"`
}

type menuResp struct {
	ID         int                `json:"id"`
	OrgID      int                `json:"org_id"`
	Name       string             `json:"name"`
	Source     string             `json:"source"`
	Categories []menuCategoryResp `json:"categories"`
}

type menuCategoryResp struct {
	ID        int            `json:"id"`
	MenuID    int            `json:"menu_id"`
	Name      string         `json:"name"`
	SortOrder int            `json:"sort_order"`
	Items     []menuItemResp `json:"items"`
}

type menuItemResp struct {
	ID          int      `json:"id"`
	CategoryID  int      `json:"category_id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Price       float64  `json:"price"`
	ImageURL    *string  `json:"image_url"`
	Tags        []string `json:"tags"`
	IsAvailable bool     `json:"is_available"`
	SortOrder   int      `json:"sort_order"`
}

func mustCreatePOS(t *testing.T, token, name string) posResp {
	t.Helper()
	body := map[string]any{
		"name":    name,
		"address": "Test street 1",
		"phone":   "+79990001122",
		"schedule": map[string]map[string]string{
			"monday": {"open": "09:00", "close": "21:00"},
		},
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/pos", body, token)
	assertStatus(t, resp, http.StatusCreated)
	var pos posResp
	decodeJSON(t, resp, &pos)
	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(), "DELETE FROM pos_locations WHERE id = $1", pos.ID)
	})
	return pos
}

func mustCreateMenu(t *testing.T, token, name string) menuResp {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/api/v1/menus", map[string]any{"name": name}, token)
	assertStatus(t, resp, http.StatusCreated)
	var menu menuResp
	decodeJSON(t, resp, &menu)
	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(), "DELETE FROM menus WHERE id = $1", menu.ID)
	})
	return menu
}

func mustAddCategory(t *testing.T, token string, menuID int, name string) menuCategoryResp {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/api/v1/menus/"+strconv.Itoa(menuID)+"/categories", map[string]any{
		"name":       name,
		"sort_order": 1,
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var cat menuCategoryResp
	decodeJSON(t, resp, &cat)
	return cat
}

func mustAddItem(t *testing.T, token string, menuID, catID int, name string) menuItemResp {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/api/v1/menus/"+strconv.Itoa(menuID)+"/categories/"+strconv.Itoa(catID)+"/items", map[string]any{
		"name":        name,
		"description": "Great dish",
		"price":       450,
		"image_url":   "https://example.com/dish.png",
		"tags":        []string{"chef", "fish"},
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var item menuItemResp
	decodeJSON(t, resp, &item)
	return item
}

func TestPOS_CRUD_AndCrossOrgProtection(t *testing.T) {
	owner := mustRegister(t, uniqueEmail(t), "password123", "POS Owner", "POS Org")
	other := mustRegister(t, uniqueEmail(t), "password123", "Other", "Other Org")

	pos := mustCreatePOS(t, owner.Tokens.AccessToken, "Main Hall")
	if pos.Name != "Main Hall" || !pos.IsActive {
		t.Fatalf("unexpected pos create response: %+v", pos)
	}

	listResp := doRequest(t, http.MethodGet, "/api/v1/pos", nil, owner.Tokens.AccessToken)
	assertStatus(t, listResp, http.StatusOK)
	var list []posResp
	decodeJSON(t, listResp, &list)
	if len(list) < 1 {
		t.Fatal("expected at least one pos location")
	}

	getResp := doRequest(t, http.MethodGet, "/api/v1/pos/"+strconv.Itoa(pos.ID), nil, owner.Tokens.AccessToken)
	assertStatus(t, getResp, http.StatusOK)
	var fetched posResp
	decodeJSON(t, getResp, &fetched)
	if fetched.ID != pos.ID {
		t.Fatalf("expected pos id %d, got %d", pos.ID, fetched.ID)
	}

	updateResp := doRequest(t, http.MethodPatch, "/api/v1/pos/"+strconv.Itoa(pos.ID), map[string]any{
		"name":      "VIP Hall",
		"is_active": false,
	}, owner.Tokens.AccessToken)
	assertStatus(t, updateResp, http.StatusOK)
	var updated posResp
	decodeJSON(t, updateResp, &updated)
	if updated.Name != "VIP Hall" || updated.IsActive {
		t.Fatalf("unexpected updated pos: %+v", updated)
	}

	forbidden := doRequest(t, http.MethodGet, "/api/v1/pos/"+strconv.Itoa(pos.ID), nil, other.Tokens.AccessToken)
	assertStatus(t, forbidden, http.StatusForbidden)
	forbidden.Body.Close()

	deleteResp := doRequest(t, http.MethodDelete, "/api/v1/pos/"+strconv.Itoa(pos.ID), nil, owner.Tokens.AccessToken)
	assertStatus(t, deleteResp, http.StatusOK)
	deleteResp.Body.Close()

	missing := doRequest(t, http.MethodGet, "/api/v1/pos/"+strconv.Itoa(pos.ID), nil, owner.Tokens.AccessToken)
	assertStatus(t, missing, http.StatusNotFound)
	missing.Body.Close()
}

func TestMenus_CreateCategoryItemUpdate_AndCrossOrgProtection(t *testing.T) {
	owner := mustRegister(t, uniqueEmail(t), "password123", "Menu Owner", "Menu Org")
	other := mustRegister(t, uniqueEmail(t), "password123", "Other", "Other Menu Org")

	menu := mustCreateMenu(t, owner.Tokens.AccessToken, "Dinner Menu")
	if menu.Source != "manual" {
		t.Fatalf("expected manual source, got %q", menu.Source)
	}

	category := mustAddCategory(t, owner.Tokens.AccessToken, menu.ID, "Hot Dishes")
	if category.MenuID != menu.ID {
		t.Fatalf("unexpected category response: %+v", category)
	}

	item := mustAddItem(t, owner.Tokens.AccessToken, menu.ID, category.ID, "Seafood Pasta")
	if item.CategoryID != category.ID || item.Name != "Seafood Pasta" {
		t.Fatalf("unexpected item response: %+v", item)
	}

	menuGetResp := doRequest(t, http.MethodGet, "/api/v1/menus/"+strconv.Itoa(menu.ID), nil, owner.Tokens.AccessToken)
	assertStatus(t, menuGetResp, http.StatusOK)
	var full menuResp
	decodeJSON(t, menuGetResp, &full)
	if len(full.Categories) != 1 || len(full.Categories[0].Items) != 1 {
		t.Fatalf("expected category+item in full menu, got %+v", full)
	}

	updateMenuResp := doRequest(t, http.MethodPatch, "/api/v1/menus/"+strconv.Itoa(menu.ID), map[string]any{"name": "Late Dinner"}, owner.Tokens.AccessToken)
	assertStatus(t, updateMenuResp, http.StatusOK)
	updateMenuResp.Body.Close()

	newPrice := 510.0
	available := false
	updateItemResp := doRequest(t, http.MethodPatch, "/api/v1/menus/items/"+strconv.Itoa(item.ID), map[string]any{
		"price":        newPrice,
		"is_available": available,
	}, owner.Tokens.AccessToken)
	assertStatus(t, updateItemResp, http.StatusOK)
	var updatedItem menuItemResp
	decodeJSON(t, updateItemResp, &updatedItem)
	if updatedItem.Price != newPrice || updatedItem.IsAvailable != available {
		t.Fatalf("unexpected updated item: %+v", updatedItem)
	}

	listResp := doRequest(t, http.MethodGet, "/api/v1/menus", nil, owner.Tokens.AccessToken)
	assertStatus(t, listResp, http.StatusOK)
	var menus []menuResp
	decodeJSON(t, listResp, &menus)
	if len(menus) < 1 {
		t.Fatal("expected at least one menu in list")
	}

	forbiddenGet := doRequest(t, http.MethodGet, "/api/v1/menus/"+strconv.Itoa(menu.ID), nil, other.Tokens.AccessToken)
	assertStatus(t, forbiddenGet, http.StatusForbidden)
	forbiddenGet.Body.Close()

	forbiddenCategory := doRequest(t, http.MethodPost, "/api/v1/menus/"+strconv.Itoa(menu.ID)+"/categories", map[string]any{"name": "Secret"}, other.Tokens.AccessToken)
	assertStatus(t, forbiddenCategory, http.StatusForbidden)
	forbiddenCategory.Body.Close()

	deleteResp := doRequest(t, http.MethodDelete, "/api/v1/menus/"+strconv.Itoa(menu.ID), nil, owner.Tokens.AccessToken)
	assertStatus(t, deleteResp, http.StatusOK)
	deleteResp.Body.Close()

	missing := doRequest(t, http.MethodGet, "/api/v1/menus/"+strconv.Itoa(menu.ID), nil, owner.Tokens.AccessToken)
	assertStatus(t, missing, http.StatusNotFound)
	missing.Body.Close()
}

func TestMenus_AddItem_RejectsMissingMenu(t *testing.T) {
	owner := mustRegister(t, uniqueEmail(t), "password123", "Menu Owner", "Menu Org")
	resp := doRequest(t, http.MethodPost, "/api/v1/menus/999999/categories/1/items", map[string]any{
		"name":  "Ghost Dish",
		"price": 100,
	}, owner.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

func TestPOS_InvalidID_ReturnsBadRequest(t *testing.T) {
	owner := mustRegister(t, uniqueEmail(t), "password123", "POS Owner", "POS Org")
	resp := doRequest(t, http.MethodGet, "/api/v1/pos/not-an-id", nil, owner.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func TestMenus_InvalidMenuID_ReturnsBadRequest(t *testing.T) {
	owner := mustRegister(t, uniqueEmail(t), "password123", "Menu Owner", "Menu Org")
	resp := doRequest(t, http.MethodGet, "/api/v1/menus/not-an-id", nil, owner.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func Example_paths() {
	fmt.Println("/api/v1/pos")
	fmt.Println("/api/v1/menus")
	// Output:
	// /api/v1/pos
	// /api/v1/menus
}
