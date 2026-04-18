package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
	posService "revisitr/internal/service/pos"
)

type Menus struct {
	pg *Module
}

func NewMenus(pg *Module) *Menus {
	return &Menus{pg: pg}
}

func (r *Menus) Create(ctx context.Context, m *entity.Menu) error {
	query := `
		INSERT INTO menus (org_id, integration_id, name, source)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return r.pg.DB().QueryRowContext(ctx, query,
		m.OrgID, m.IntegrationID, m.Name, m.Source,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Menus) GetByID(ctx context.Context, id int) (*entity.Menu, error) {
	var m entity.Menu
	err := r.pg.DB().GetContext(ctx, &m, "SELECT * FROM menus WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("menus.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("menus.GetByID: %w", err)
	}
	return &m, nil
}

func (r *Menus) GetByOrgID(ctx context.Context, orgID int) ([]entity.Menu, error) {
	var menus []entity.Menu
	err := r.pg.DB().SelectContext(ctx, &menus,
		"SELECT * FROM menus WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("menus.GetByOrgID: %w", err)
	}
	for i := range menus {
		bindings, bindErr := r.GetMenuPOSBindings(ctx, menus[i].ID)
		if bindErr != nil {
			return nil, bindErr
		}
		menus[i].Bindings = bindings
	}
	return menus, nil
}

// GetByOrgAndIntegration returns the POS-imported menu for an org+integration pair.
func (r *Menus) GetByOrgAndIntegration(ctx context.Context, orgID, integrationID int) (*entity.Menu, error) {
	var m entity.Menu
	err := r.pg.DB().GetContext(ctx, &m,
		"SELECT * FROM menus WHERE org_id = $1 AND integration_id = $2 AND source = 'pos_import' LIMIT 1",
		orgID, integrationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("menus.GetByOrgAndIntegration: %w", err)
	}
	return &m, nil
}

// UpsertFromPOS creates or updates a menu from POS provider data.
// It only overwrites name and price for existing items, preserving
// manually edited description, tags, and image_url fields.
func (r *Menus) UpsertFromPOS(ctx context.Context, integrationID, orgID int, posMenu *posService.POSMenu) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("menus.UpsertFromPOS begin: %w", err)
	}
	defer tx.Rollback()

	// Find or create the POS-imported menu for this integration.
	var menuID int
	err = tx.QueryRowContext(ctx,
		"SELECT id FROM menus WHERE org_id = $1 AND integration_id = $2 AND source = 'pos_import' LIMIT 1",
		orgID, integrationID).Scan(&menuID)
	if err == sql.ErrNoRows {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO menus (org_id, integration_id, name, source)
			 VALUES ($1, $2, $3, 'pos_import')
			 RETURNING id`,
			orgID, integrationID, "POS Menu").Scan(&menuID)
		if err != nil {
			return fmt.Errorf("menus.UpsertFromPOS create menu: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("menus.UpsertFromPOS find menu: %w", err)
	}

	for catIdx, posCat := range posMenu.Categories {
		// Find or create category by name within this menu.
		var catID int
		err = tx.QueryRowContext(ctx,
			"SELECT id FROM menu_categories WHERE menu_id = $1 AND name = $2 LIMIT 1",
			menuID, posCat.Name).Scan(&catID)
		if err == sql.ErrNoRows {
			err = tx.QueryRowContext(ctx,
				`INSERT INTO menu_categories (menu_id, name, sort_order)
					 VALUES ($1, $2, $3) RETURNING id`,
				menuID, posCat.Name, catIdx).Scan(&catID)
			if err != nil {
				return fmt.Errorf("menus.UpsertFromPOS create category %q: %w", posCat.Name, err)
			}
		} else if err != nil {
			return fmt.Errorf("menus.UpsertFromPOS find category: %w", err)
		}

		for itemIdx, posItem := range posCat.Items {
			if posItem.ExternalID == "" {
				continue
			}
			// Find existing item by external_id within this category's menu.
			var itemID int
			err = tx.QueryRowContext(ctx,
				`SELECT mi.id FROM menu_items mi
				 JOIN menu_categories mc ON mc.id = mi.category_id
				 WHERE mc.menu_id = $1 AND mi.external_id = $2 LIMIT 1`,
				menuID, posItem.ExternalID).Scan(&itemID)

			if err == sql.ErrNoRows {
				// Create new item.
				_, err = tx.ExecContext(ctx,
					`INSERT INTO menu_items (category_id, name, price, external_id, sort_order, is_available)
					 VALUES ($1, $2, $3, $4, $5, true)`,
					catID, posItem.Name, posItem.Price, posItem.ExternalID, itemIdx)
				if err != nil {
					return fmt.Errorf("menus.UpsertFromPOS create item %q: %w", posItem.Name, err)
				}
			} else if err != nil {
				return fmt.Errorf("menus.UpsertFromPOS find item: %w", err)
			} else {
				// Update only name and price; preserve description, tags, image_url.
				_, err = tx.ExecContext(ctx,
					`UPDATE menu_items SET name = $1, price = $2, category_id = $3, updated_at = NOW()
					 WHERE id = $4`,
					posItem.Name, posItem.Price, catID, itemID)
				if err != nil {
					return fmt.Errorf("menus.UpsertFromPOS update item %q: %w", posItem.Name, err)
				}
			}
		}
	}

	// Update last_synced_at on the menu.
	_, err = tx.ExecContext(ctx,
		"UPDATE menus SET last_synced_at = NOW(), updated_at = NOW() WHERE id = $1", menuID)
	if err != nil {
		return fmt.Errorf("menus.UpsertFromPOS update last_synced_at: %w", err)
	}

	return tx.Commit()
}

func (r *Menus) Update(ctx context.Context, m *entity.Menu) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE menus SET name = $1, intro_content = $2, updated_at = NOW() WHERE id = $3",
		m.Name, m.IntroContent, m.ID)
	if err != nil {
		return fmt.Errorf("menus.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("menus.Update: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Menus) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM menus WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("menus.Delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("menus.Delete: %w", sql.ErrNoRows)
	}
	return nil
}

// Categories

func (r *Menus) CreateCategory(ctx context.Context, cat *entity.MenuCategory) error {
	query := `
		INSERT INTO menu_categories (menu_id, name, icon_emoji, icon_image_url, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pg.DB().QueryRowContext(ctx, query,
		cat.MenuID, cat.Name, cat.IconEmoji, cat.IconImageURL, cat.SortOrder,
	).Scan(&cat.ID, &cat.CreatedAt)
}

func (r *Menus) GetCategories(ctx context.Context, menuID int) ([]entity.MenuCategory, error) {
	var cats []entity.MenuCategory
	err := r.pg.DB().SelectContext(ctx, &cats,
		"SELECT * FROM menu_categories WHERE menu_id = $1 ORDER BY sort_order", menuID)
	if err != nil {
		return nil, fmt.Errorf("menus.GetCategories: %w", err)
	}
	return cats, nil
}

func (r *Menus) DeleteCategory(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM menu_categories WHERE id = $1", id)
	return err
}

func (r *Menus) GetCategory(ctx context.Context, id int) (*entity.MenuCategory, error) {
	var category entity.MenuCategory
	err := r.pg.DB().GetContext(ctx, &category, "SELECT * FROM menu_categories WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("menus.GetCategory: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("menus.GetCategory: %w", err)
	}
	return &category, nil
}

func (r *Menus) UpdateCategory(ctx context.Context, category *entity.MenuCategory) error {
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE menu_categories
		SET name = $1, icon_emoji = $2, icon_image_url = $3, sort_order = $4
		WHERE id = $5`,
		category.Name, category.IconEmoji, category.IconImageURL, category.SortOrder, category.ID)
	if err != nil {
		return fmt.Errorf("menus.UpdateCategory: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("menus.UpdateCategory: %w", sql.ErrNoRows)
	}
	return nil
}

// Items

func (r *Menus) CreateItem(ctx context.Context, item *entity.MenuItem) error {
	tagsVal, _ := item.Tags.Value()
	query := `
		INSERT INTO menu_items (category_id, name, description, price, weight, image_url, tags, external_id, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`
	return r.pg.DB().QueryRowContext(ctx, query,
		item.CategoryID, item.Name, item.Description, item.Price, item.Weight,
		item.ImageURL, tagsVal, item.ExternalID, item.SortOrder,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *Menus) GetItemsByCategory(ctx context.Context, categoryID int) ([]entity.MenuItem, error) {
	var items []entity.MenuItem
	err := r.pg.DB().SelectContext(ctx, &items,
		"SELECT * FROM menu_items WHERE category_id = $1 ORDER BY sort_order", categoryID)
	if err != nil {
		return nil, fmt.Errorf("menus.GetItemsByCategory: %w", err)
	}
	return items, nil
}

func (r *Menus) GetItem(ctx context.Context, id int) (*entity.MenuItem, error) {
	var item entity.MenuItem
	err := r.pg.DB().GetContext(ctx, &item, "SELECT * FROM menu_items WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("menus.GetItem: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("menus.GetItem: %w", err)
	}
	return &item, nil
}

func (r *Menus) UpdateItem(ctx context.Context, item *entity.MenuItem) error {
	tagsVal, _ := item.Tags.Value()
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE menu_items
		SET name = $1, description = $2, price = $3, weight = $4, image_url = $5,
		    tags = $6, is_available = $7, sort_order = $8, updated_at = NOW()
		WHERE id = $9`,
		item.Name, item.Description, item.Price, item.Weight, item.ImageURL,
		tagsVal, item.IsAvailable, item.SortOrder, item.ID)
	if err != nil {
		return fmt.Errorf("menus.UpdateItem: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("menus.UpdateItem: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Menus) DeleteItem(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM menu_items WHERE id = $1", id)
	return err
}

// GetFullMenu loads a menu with all categories and items.
func (r *Menus) GetFullMenu(ctx context.Context, menuID int) (*entity.Menu, error) {
	menu, err := r.GetByID(ctx, menuID)
	if err != nil {
		return nil, err
	}

	cats, err := r.GetCategories(ctx, menuID)
	if err != nil {
		return nil, err
	}

	for i := range cats {
		items, err := r.GetItemsByCategory(ctx, cats[i].ID)
		if err != nil {
			return nil, err
		}
		cats[i].Items = items
	}

	menu.Categories = cats
	bindings, err := r.GetMenuPOSBindings(ctx, menuID)
	if err != nil {
		return nil, err
	}
	menu.Bindings = bindings
	return menu, nil
}

// GetClientOrderStats aggregates order data for a specific client.
func (r *Menus) GetClientOrderStats(ctx context.Context, clientID int) (*entity.ClientOrderStats, error) {
	stats := &entity.ClientOrderStats{}
	err := r.pg.DB().GetContext(ctx, stats, `
		SELECT
			COUNT(*) AS total_orders,
			COALESCE(SUM(total), 0) AS total_amount,
			CASE WHEN COUNT(*) > 0 THEN SUM(total) / COUNT(*) ELSE 0 END AS avg_amount,
			MAX(ordered_at) AS last_order_at
		FROM external_orders
		WHERE client_id = $1`, clientID)
	if err != nil {
		return nil, fmt.Errorf("menus.GetClientOrderStats: %w", err)
	}

	var topItems []entity.TopOrderItem
	err = r.pg.DB().SelectContext(ctx, &topItems, `
		SELECT
			item->>'name' AS name,
			COUNT(DISTINCT eo.id) AS order_count,
			SUM((item->>'quantity')::INT) AS total_qty,
			SUM((item->>'price')::NUMERIC * (item->>'quantity')::INT) AS total_sum
		FROM external_orders eo,
		     jsonb_array_elements(eo.items) AS item
		WHERE eo.client_id = $1
		GROUP BY item->>'name'
		ORDER BY order_count DESC
		LIMIT 10`, clientID)
	if err == nil {
		stats.TopItems = topItems
	}

	return stats, nil
}

// Bot-POS locations

func (r *Menus) SetBotPOSLocations(ctx context.Context, botID int, posIDs []int) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("menus.SetBotPOSLocations begin: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM bot_pos_locations WHERE bot_id = $1", botID)
	if err != nil {
		return fmt.Errorf("menus.SetBotPOSLocations delete: %w", err)
	}

	for _, posID := range posIDs {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO bot_pos_locations (bot_id, pos_id) VALUES ($1, $2)", botID, posID)
		if err != nil {
			return fmt.Errorf("menus.SetBotPOSLocations insert: %w", err)
		}
	}

	if len(posIDs) == 0 {
		if _, err := tx.ExecContext(ctx,
			"UPDATE bots SET status = 'pending', updated_at = NOW() WHERE id = $1", botID); err != nil {
			return fmt.Errorf("menus.SetBotPOSLocations set pending: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Menus) GetBotPOSLocations(ctx context.Context, botID int) ([]int, error) {
	var posIDs []int
	err := r.pg.DB().SelectContext(ctx, &posIDs,
		"SELECT pos_id FROM bot_pos_locations WHERE bot_id = $1 ORDER BY pos_id", botID)
	if err != nil {
		return nil, fmt.Errorf("menus.GetBotPOSLocations: %w", err)
	}
	return posIDs, nil
}

func (r *Menus) SetMenuPOSBindings(ctx context.Context, menuID int, bindings []entity.MenuPOSBindingRequest) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("menus.SetMenuPOSBindings begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM menu_pos_bindings WHERE menu_id = $1", menuID); err != nil {
		return fmt.Errorf("menus.SetMenuPOSBindings delete: %w", err)
	}

	for _, binding := range bindings {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO menu_pos_bindings (menu_id, pos_id, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())`,
			menuID, binding.POSID, binding.IsActive); err != nil {
			return fmt.Errorf("menus.SetMenuPOSBindings insert: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Menus) GetMenuPOSBindings(ctx context.Context, menuID int) ([]entity.MenuPOSBinding, error) {
	var bindings []entity.MenuPOSBinding
	err := r.pg.DB().SelectContext(ctx, &bindings, `
		SELECT mpb.menu_id, mpb.pos_id, mpb.is_active, mpb.created_at, mpb.updated_at,
		       COALESCE(pl.name, '') AS pos_name
		FROM menu_pos_bindings mpb
		LEFT JOIN pos_locations pl ON pl.id = mpb.pos_id
		WHERE mpb.menu_id = $1
		ORDER BY mpb.created_at ASC, mpb.pos_id ASC`, menuID)
	if err != nil {
		return nil, fmt.Errorf("menus.GetMenuPOSBindings: %w", err)
	}
	return bindings, nil
}

func (r *Menus) GetActiveMenuForPOS(ctx context.Context, orgID, posID int) (*entity.Menu, error) {
	var menu entity.Menu
	err := r.pg.DB().GetContext(ctx, &menu, `
		SELECT m.*
		FROM menu_pos_bindings mpb
		JOIN menus m ON m.id = mpb.menu_id
		WHERE m.org_id = $1 AND mpb.pos_id = $2 AND mpb.is_active = true
		ORDER BY mpb.created_at ASC, m.id ASC
		LIMIT 1`, orgID, posID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("menus.GetActiveMenuForPOS: %w", err)
	}
	return r.GetFullMenu(ctx, menu.ID)
}
