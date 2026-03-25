package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Marketplace struct {
	pg *Module
}

func NewMarketplace(pg *Module) *Marketplace {
	return &Marketplace{pg: pg}
}

// ── Products ─────────────────────────────────────────────────────────────────

func (r *Marketplace) CreateProduct(ctx context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error) {
	var p entity.MarketplaceProduct
	err := r.pg.DB().GetContext(ctx, &p, `
		INSERT INTO marketplace_products (org_id, name, description, image_url, price_points, stock, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *`,
		orgID, req.Name, req.Description, req.ImageURL, req.PricePoints, req.Stock, req.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("marketplace.CreateProduct: %w", err)
	}
	return &p, nil
}

func (r *Marketplace) GetProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error) {
	var products []entity.MarketplaceProduct
	err := r.pg.DB().SelectContext(ctx, &products,
		"SELECT * FROM marketplace_products WHERE org_id = $1 ORDER BY sort_order, name", orgID)
	if err != nil {
		return nil, fmt.Errorf("marketplace.GetProducts: %w", err)
	}
	return products, nil
}

func (r *Marketplace) GetActiveProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error) {
	var products []entity.MarketplaceProduct
	err := r.pg.DB().SelectContext(ctx, &products,
		"SELECT * FROM marketplace_products WHERE org_id = $1 AND is_active = true ORDER BY sort_order, name", orgID)
	if err != nil {
		return nil, fmt.Errorf("marketplace.GetActiveProducts: %w", err)
	}
	return products, nil
}

func (r *Marketplace) GetProductByID(ctx context.Context, id int) (*entity.MarketplaceProduct, error) {
	var p entity.MarketplaceProduct
	err := r.pg.DB().GetContext(ctx, &p,
		"SELECT * FROM marketplace_products WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("marketplace.GetProductByID: %w", err)
	}
	return &p, nil
}

func (r *Marketplace) UpdateProduct(ctx context.Context, id int, req entity.UpdateProductRequest) (*entity.MarketplaceProduct, error) {
	var p entity.MarketplaceProduct
	err := r.pg.DB().GetContext(ctx, &p, `
		UPDATE marketplace_products SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			image_url = COALESCE($3, image_url),
			price_points = COALESCE($4, price_points),
			stock = COALESCE($5, stock),
			is_active = COALESCE($6, is_active),
			sort_order = COALESCE($7, sort_order),
			updated_at = now()
		WHERE id = $8
		RETURNING *`,
		req.Name, req.Description, req.ImageURL, req.PricePoints,
		req.Stock, req.IsActive, req.SortOrder, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("marketplace.UpdateProduct: %w", err)
	}
	return &p, nil
}

func (r *Marketplace) DeleteProduct(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM marketplace_products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("marketplace.DeleteProduct: %w", err)
	}
	return nil
}

func (r *Marketplace) DecrementStock(ctx context.Context, productID int, qty int) error {
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE marketplace_products SET stock = stock - $1, updated_at = now()
		WHERE id = $2 AND stock IS NOT NULL AND stock >= $1`,
		qty, productID)
	if err != nil {
		return fmt.Errorf("marketplace.DecrementStock: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("marketplace.DecrementStock: insufficient stock")
	}
	return nil
}

// ── Orders ───────────────────────────────────────────────────────────────────

func (r *Marketplace) CreateOrder(ctx context.Context, order *entity.MarketplaceOrder) error {
	err := r.pg.DB().GetContext(ctx, order, `
		INSERT INTO marketplace_orders (org_id, client_id, status, total_points, items, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`,
		order.OrgID, order.ClientID, order.Status, order.TotalPoints, order.Items, order.Note)
	if err != nil {
		return fmt.Errorf("marketplace.CreateOrder: %w", err)
	}
	return nil
}

func (r *Marketplace) GetOrders(ctx context.Context, orgID int) ([]entity.MarketplaceOrder, error) {
	var orders []entity.MarketplaceOrder
	err := r.pg.DB().SelectContext(ctx, &orders,
		"SELECT * FROM marketplace_orders WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("marketplace.GetOrders: %w", err)
	}
	return orders, nil
}

func (r *Marketplace) GetOrderByID(ctx context.Context, id int) (*entity.MarketplaceOrder, error) {
	var o entity.MarketplaceOrder
	err := r.pg.DB().GetContext(ctx, &o,
		"SELECT * FROM marketplace_orders WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("marketplace.GetOrderByID: %w", err)
	}
	return &o, nil
}

func (r *Marketplace) GetOrdersByClient(ctx context.Context, clientID int) ([]entity.MarketplaceOrder, error) {
	var orders []entity.MarketplaceOrder
	err := r.pg.DB().SelectContext(ctx, &orders,
		"SELECT * FROM marketplace_orders WHERE client_id = $1 ORDER BY created_at DESC", clientID)
	if err != nil {
		return nil, fmt.Errorf("marketplace.GetOrdersByClient: %w", err)
	}
	return orders, nil
}

func (r *Marketplace) UpdateOrderStatus(ctx context.Context, id int, status string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"UPDATE marketplace_orders SET status = $1, updated_at = now() WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("marketplace.UpdateOrderStatus: %w", err)
	}
	return nil
}

func (r *Marketplace) GetStats(ctx context.Context, orgID int) (*entity.MarketplaceStats, error) {
	var stats entity.MarketplaceStats
	err := r.pg.DB().GetContext(ctx, &stats, `
		SELECT
			(SELECT COUNT(*) FROM marketplace_products WHERE org_id = $1) AS total_products,
			(SELECT COUNT(*) FROM marketplace_products WHERE org_id = $1 AND is_active = true) AS active_products,
			(SELECT COUNT(*) FROM marketplace_orders WHERE org_id = $1) AS total_orders,
			(SELECT COALESCE(SUM(total_points), 0) FROM marketplace_orders WHERE org_id = $1 AND status != 'cancelled') AS total_spent_points`,
		orgID)
	if err != nil {
		return nil, fmt.Errorf("marketplace.GetStats: %w", err)
	}
	return &stats, nil
}
