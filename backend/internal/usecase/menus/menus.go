package menus

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrNotFound = errors.New("menu not found")
	ErrNotOwner = errors.New("not the owner of this menu")
)

type menusRepo interface {
	Create(ctx context.Context, m *entity.Menu) error
	GetByID(ctx context.Context, id int) (*entity.Menu, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Menu, error)
	Update(ctx context.Context, m *entity.Menu) error
	Delete(ctx context.Context, id int) error
	GetFullMenu(ctx context.Context, menuID int) (*entity.Menu, error)
	CreateCategory(ctx context.Context, cat *entity.MenuCategory) error
	GetCategory(ctx context.Context, id int) (*entity.MenuCategory, error)
	GetCategories(ctx context.Context, menuID int) ([]entity.MenuCategory, error)
	UpdateCategory(ctx context.Context, category *entity.MenuCategory) error
	DeleteCategory(ctx context.Context, id int) error
	CreateItem(ctx context.Context, item *entity.MenuItem) error
	GetItem(ctx context.Context, id int) (*entity.MenuItem, error)
	UpdateItem(ctx context.Context, item *entity.MenuItem) error
	DeleteItem(ctx context.Context, id int) error
	GetClientOrderStats(ctx context.Context, clientID int) (*entity.ClientOrderStats, error)
	SetBotPOSLocations(ctx context.Context, botID int, posIDs []int) error
	GetBotPOSLocations(ctx context.Context, botID int) ([]int, error)
	SetMenuPOSBindings(ctx context.Context, menuID int, bindings []entity.MenuPOSBindingRequest) error
	GetMenuPOSBindings(ctx context.Context, menuID int) ([]entity.MenuPOSBinding, error)
	GetActiveMenuForPOS(ctx context.Context, orgID, posID int) (*entity.Menu, error)
}

type Usecase struct {
	logger *slog.Logger
	repo   menusRepo
}

func New(repo menusRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Create(ctx context.Context, orgID int, req entity.CreateMenuRequest) (*entity.Menu, error) {
	m := &entity.Menu{
		OrgID:  orgID,
		Name:   req.Name,
		Source: "manual",
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return m, nil
}

func (uc *Usecase) Get(ctx context.Context, orgID, menuID int) (*entity.Menu, error) {
	m, err := uc.repo.GetFullMenu(ctx, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if m.OrgID != orgID {
		return nil, ErrNotOwner
	}
	return m, nil
}

func (uc *Usecase) List(ctx context.Context, orgID int) ([]entity.Menu, error) {
	return uc.repo.GetByOrgID(ctx, orgID)
}

func (uc *Usecase) Update(ctx context.Context, orgID, menuID int, req entity.UpdateMenuRequest) error {
	m, err := uc.repo.GetByID(ctx, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if m.OrgID != orgID {
		return ErrNotOwner
	}

	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.IntroContent != nil {
		if err := req.IntroContent.Validate(); err != nil {
			return fmt.Errorf("invalid intro content: %w", err)
		}
		m.IntroContent = req.IntroContent
	}
	if err := uc.repo.Update(ctx, m); err != nil {
		return err
	}
	if req.Bindings != nil {
		if err := uc.repo.SetMenuPOSBindings(ctx, menuID, req.Bindings); err != nil {
			return err
		}
	}
	return nil
}

func (uc *Usecase) Delete(ctx context.Context, orgID, menuID int) error {
	m, err := uc.repo.GetByID(ctx, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if m.OrgID != orgID {
		return ErrNotOwner
	}
	return uc.repo.Delete(ctx, menuID)
}

func (uc *Usecase) AddCategory(ctx context.Context, orgID, menuID int, req entity.CreateMenuCategoryRequest) (*entity.MenuCategory, error) {
	m, err := uc.repo.GetByID(ctx, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if m.OrgID != orgID {
		return nil, ErrNotOwner
	}

	cat := &entity.MenuCategory{
		MenuID:       menuID,
		Name:         req.Name,
		SortOrder:    req.SortOrder,
		IconEmoji:    emptyStringPtr(req.IconEmoji),
		IconImageURL: emptyStringPtr(req.IconImageURL),
	}
	if err := uc.repo.CreateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (uc *Usecase) UpdateCategory(ctx context.Context, orgID, categoryID int, req entity.UpdateMenuCategoryRequest) (*entity.MenuCategory, error) {
	category, err := uc.repo.GetCategory(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	menu, err := uc.repo.GetByID(ctx, category.MenuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if menu.OrgID != orgID {
		return nil, ErrNotOwner
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.IconEmoji != nil {
		category.IconEmoji = req.IconEmoji
	}
	if req.IconImageURL != nil {
		category.IconImageURL = req.IconImageURL
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if err := uc.repo.UpdateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (uc *Usecase) AddItem(ctx context.Context, orgID, menuID, categoryID int, req entity.CreateMenuItemRequest) (*entity.MenuItem, error) {
	m, err := uc.repo.GetByID(ctx, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if m.OrgID != orgID {
		return nil, ErrNotOwner
	}

	item := &entity.MenuItem{
		CategoryID:  categoryID,
		Name:        req.Name,
		Description: &req.Description,
		Price:       req.Price,
		Weight:      emptyStringPtr(req.Weight),
		ImageURL:    &req.ImageURL,
		Tags:        req.Tags,
		IsAvailable: true,
	}
	if err := uc.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *Usecase) UpdateItem(ctx context.Context, orgID, itemID int, req entity.UpdateMenuItemRequest) (*entity.MenuItem, error) {
	item, err := uc.repo.GetItem(ctx, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	category, err := uc.repo.GetCategory(ctx, item.CategoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	menu, err := uc.repo.GetByID(ctx, category.MenuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if menu.OrgID != orgID {
		return nil, ErrNotOwner
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Weight != nil {
		item.Weight = req.Weight
	}
	if req.ImageURL != nil {
		item.ImageURL = req.ImageURL
	}
	if req.Tags != nil {
		item.Tags = *req.Tags
	}
	if req.IsAvailable != nil {
		item.IsAvailable = *req.IsAvailable
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}

	if err := uc.repo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *Usecase) GetClientOrderStats(ctx context.Context, clientID int) (*entity.ClientOrderStats, error) {
	return uc.repo.GetClientOrderStats(ctx, clientID)
}

func (uc *Usecase) SetBotPOSLocations(ctx context.Context, botID int, posIDs []int) error {
	return uc.repo.SetBotPOSLocations(ctx, botID, posIDs)
}

func (uc *Usecase) GetBotPOSLocations(ctx context.Context, botID int) ([]int, error) {
	return uc.repo.GetBotPOSLocations(ctx, botID)
}

func (uc *Usecase) GetActiveMenuForPOS(ctx context.Context, orgID, posID int) (*entity.Menu, error) {
	return uc.repo.GetActiveMenuForPOS(ctx, orgID, posID)
}

func emptyStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
