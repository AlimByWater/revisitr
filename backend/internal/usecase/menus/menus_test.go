package menus

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockMenusRepo struct {
	createFn             func(ctx context.Context, m *entity.Menu) error
	getByIDFn            func(ctx context.Context, id int) (*entity.Menu, error)
	getByOrgIDFn         func(ctx context.Context, orgID int) ([]entity.Menu, error)
	updateFn             func(ctx context.Context, m *entity.Menu) error
	deleteFn             func(ctx context.Context, id int) error
	getFullMenuFn        func(ctx context.Context, menuID int) (*entity.Menu, error)
	createCategoryFn     func(ctx context.Context, cat *entity.MenuCategory) error
	getCategoriesFn      func(ctx context.Context, menuID int) ([]entity.MenuCategory, error)
	deleteCategoryFn     func(ctx context.Context, id int) error
	createItemFn         func(ctx context.Context, item *entity.MenuItem) error
	getItemFn            func(ctx context.Context, id int) (*entity.MenuItem, error)
	updateItemFn         func(ctx context.Context, item *entity.MenuItem) error
	deleteItemFn         func(ctx context.Context, id int) error
	getClientOrderStatsFn func(ctx context.Context, clientID int) (*entity.ClientOrderStats, error)
	setBotPOSLocationsFn  func(ctx context.Context, botID int, posIDs []int) error
	getBotPOSLocationsFn  func(ctx context.Context, botID int) ([]int, error)
}

func (m *mockMenusRepo) Create(ctx context.Context, menu *entity.Menu) error {
	if m.createFn != nil {
		return m.createFn(ctx, menu)
	}
	return nil
}
func (m *mockMenusRepo) GetByID(ctx context.Context, id int) (*entity.Menu, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockMenusRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Menu, error) {
	if m.getByOrgIDFn != nil {
		return m.getByOrgIDFn(ctx, orgID)
	}
	return nil, nil
}
func (m *mockMenusRepo) Update(ctx context.Context, menu *entity.Menu) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, menu)
	}
	return nil
}
func (m *mockMenusRepo) Delete(ctx context.Context, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockMenusRepo) GetFullMenu(ctx context.Context, menuID int) (*entity.Menu, error) {
	if m.getFullMenuFn != nil {
		return m.getFullMenuFn(ctx, menuID)
	}
	return nil, nil
}
func (m *mockMenusRepo) CreateCategory(ctx context.Context, cat *entity.MenuCategory) error {
	if m.createCategoryFn != nil {
		return m.createCategoryFn(ctx, cat)
	}
	return nil
}
func (m *mockMenusRepo) GetCategories(ctx context.Context, menuID int) ([]entity.MenuCategory, error) {
	if m.getCategoriesFn != nil {
		return m.getCategoriesFn(ctx, menuID)
	}
	return nil, nil
}
func (m *mockMenusRepo) DeleteCategory(ctx context.Context, id int) error {
	if m.deleteCategoryFn != nil {
		return m.deleteCategoryFn(ctx, id)
	}
	return nil
}
func (m *mockMenusRepo) CreateItem(ctx context.Context, item *entity.MenuItem) error {
	if m.createItemFn != nil {
		return m.createItemFn(ctx, item)
	}
	return nil
}
func (m *mockMenusRepo) GetItem(ctx context.Context, id int) (*entity.MenuItem, error) {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, id)
	}
	return nil, nil
}
func (m *mockMenusRepo) UpdateItem(ctx context.Context, item *entity.MenuItem) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, item)
	}
	return nil
}
func (m *mockMenusRepo) DeleteItem(ctx context.Context, id int) error {
	if m.deleteItemFn != nil {
		return m.deleteItemFn(ctx, id)
	}
	return nil
}
func (m *mockMenusRepo) GetClientOrderStats(ctx context.Context, clientID int) (*entity.ClientOrderStats, error) {
	if m.getClientOrderStatsFn != nil {
		return m.getClientOrderStatsFn(ctx, clientID)
	}
	return nil, nil
}
func (m *mockMenusRepo) SetBotPOSLocations(ctx context.Context, botID int, posIDs []int) error {
	if m.setBotPOSLocationsFn != nil {
		return m.setBotPOSLocationsFn(ctx, botID, posIDs)
	}
	return nil
}
func (m *mockMenusRepo) GetBotPOSLocations(ctx context.Context, botID int) ([]int, error) {
	if m.getBotPOSLocationsFn != nil {
		return m.getBotPOSLocationsFn(ctx, botID)
	}
	return nil, nil
}

// --- tests ---

func TestCreate(t *testing.T) {
	var createdMenu *entity.Menu
	repo := &mockMenusRepo{
		createFn: func(_ context.Context, m *entity.Menu) error {
			m.ID = 1
			createdMenu = m
			return nil
		},
	}
	uc := New(repo)

	result, err := uc.Create(context.Background(), 10, entity.CreateMenuRequest{
		Name: "Main Menu",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OrgID != 10 {
		t.Errorf("expected org_id=10, got %d", result.OrgID)
	}
	if result.Name != "Main Menu" {
		t.Errorf("expected name='Main Menu', got '%s'", result.Name)
	}
	if result.Source != "manual" {
		t.Errorf("expected source='manual', got '%s'", result.Source)
	}
	if createdMenu == nil {
		t.Fatal("expected repo.Create to be called")
	}
}

func TestGet_Success(t *testing.T) {
	repo := &mockMenusRepo{
		getFullMenuFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return &entity.Menu{
				ID:    1,
				OrgID: 10,
				Name:  "Main Menu",
				Categories: []entity.MenuCategory{
					{ID: 1, MenuID: 1, Name: "Drinks"},
				},
			}, nil
		},
	}
	uc := New(repo)

	result, err := uc.Get(context.Background(), 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Main Menu" {
		t.Errorf("expected name='Main Menu', got '%s'", result.Name)
	}
	if len(result.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(result.Categories))
	}
}

func TestGet_NotFound(t *testing.T) {
	repo := &mockMenusRepo{
		getFullMenuFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return nil, fmt.Errorf("menus.GetFullMenu: %w", sql.ErrNoRows)
		},
	}
	uc := New(repo)

	_, err := uc.Get(context.Background(), 10, 999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestGet_NotOwner(t *testing.T) {
	repo := &mockMenusRepo{
		getFullMenuFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return &entity.Menu{ID: 1, OrgID: 5, Name: "Other Org Menu"}, nil
		},
	}
	uc := New(repo)

	_, err := uc.Get(context.Background(), 99, 1) // caller orgID=99, menu orgID=5
	if err != ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got: %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	deleteCalled := false
	repo := &mockMenusRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return &entity.Menu{ID: 1, OrgID: 10}, nil
		},
		deleteFn: func(_ context.Context, id int) error {
			deleteCalled = true
			if id != 1 {
				t.Errorf("expected delete id=1, got %d", id)
			}
			return nil
		},
	}
	uc := New(repo)

	if err := uc.Delete(context.Background(), 10, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected repo.Delete to be called")
	}
}

func TestDelete_NotOwner(t *testing.T) {
	repo := &mockMenusRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return &entity.Menu{ID: 1, OrgID: 5}, nil
		},
	}
	uc := New(repo)

	err := uc.Delete(context.Background(), 99, 1) // caller orgID=99, menu orgID=5
	if err != ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got: %v", err)
	}
}

func TestAddCategory(t *testing.T) {
	var createdCat *entity.MenuCategory
	repo := &mockMenusRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return &entity.Menu{ID: 1, OrgID: 10}, nil
		},
		createCategoryFn: func(_ context.Context, cat *entity.MenuCategory) error {
			cat.ID = 5
			createdCat = cat
			return nil
		},
	}
	uc := New(repo)

	result, err := uc.AddCategory(context.Background(), 10, 1, entity.CreateMenuCategoryRequest{
		Name:      "Desserts",
		SortOrder: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Desserts" {
		t.Errorf("expected name='Desserts', got '%s'", result.Name)
	}
	if result.MenuID != 1 {
		t.Errorf("expected menu_id=1, got %d", result.MenuID)
	}
	if result.SortOrder != 3 {
		t.Errorf("expected sort_order=3, got %d", result.SortOrder)
	}
	if createdCat == nil {
		t.Fatal("expected repo.CreateCategory to be called")
	}
}

func TestAddItem(t *testing.T) {
	var createdItem *entity.MenuItem
	repo := &mockMenusRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Menu, error) {
			return &entity.Menu{ID: 1, OrgID: 10}, nil
		},
		createItemFn: func(_ context.Context, item *entity.MenuItem) error {
			item.ID = 7
			createdItem = item
			return nil
		},
	}
	uc := New(repo)

	result, err := uc.AddItem(context.Background(), 10, 1, 5, entity.CreateMenuItemRequest{
		Name:        "Cheesecake",
		Description: "Classic New York style",
		Price:       450.0,
		ImageURL:    "https://example.com/cheesecake.jpg",
		Tags:        entity.Tags{"dessert", "popular"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Cheesecake" {
		t.Errorf("expected name='Cheesecake', got '%s'", result.Name)
	}
	if result.CategoryID != 5 {
		t.Errorf("expected category_id=5, got %d", result.CategoryID)
	}
	if result.Price != 450.0 {
		t.Errorf("expected price=450.0, got %f", result.Price)
	}
	if !result.IsAvailable {
		t.Error("expected is_available=true by default")
	}
	if createdItem == nil {
		t.Fatal("expected repo.CreateItem to be called")
	}
}
