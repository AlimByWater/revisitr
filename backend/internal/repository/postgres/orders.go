package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Orders struct {
	pg *Module
}

func NewOrders(pg *Module) *Orders {
	return &Orders{pg: pg}
}

func (r *Orders) Create(ctx context.Context, o *entity.Order) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("orders.Create begin: %w", err)
	}
	defer tx.Rollback()

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO orders (bot_id, bot_client_id, source, format_id, format_name, table_num, total_price, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`,
		o.BotID, o.BotClientID, o.Source, o.FormatID, o.FormatName, o.TableNum, o.TotalPrice, o.Status,
	).Scan(&o.ID, &o.CreatedAt); err != nil {
		return fmt.Errorf("orders.Create insert: %w", err)
	}

	for i := range o.Items {
		item := &o.Items[i]
		item.OrderID = o.ID
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO order_items (order_id, course_id, course_title, menu_item_id, item_name, price, surcharge)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			item.OrderID, item.CourseID, item.CourseTitle, item.MenuItemID,
			item.ItemName, item.Price, item.Surcharge,
		).Scan(&item.ID); err != nil {
			return fmt.Errorf("orders.Create item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("orders.Create commit: %w", err)
	}
	return nil
}

func (r *Orders) ListByBot(ctx context.Context, botID int, source, status string) ([]entity.Order, error) {
	query := "SELECT * FROM orders WHERE bot_id = $1"
	args := []any{botID}
	if source != "" {
		args = append(args, source)
		query += fmt.Sprintf(" AND source = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}
	query += " ORDER BY created_at DESC"

	var orders []entity.Order
	if err := r.pg.DB().SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, fmt.Errorf("orders.ListByBot: %w", err)
	}
	for i := range orders {
		if err := r.pg.DB().SelectContext(ctx, &orders[i].Items,
			"SELECT * FROM order_items WHERE order_id = $1 ORDER BY id",
			orders[i].ID); err != nil {
			return nil, fmt.Errorf("orders.ListByBot items: %w", err)
		}
	}
	return orders, nil
}

func (r *Orders) ListByOrg(ctx context.Context, orgID int, source, status string) ([]entity.Order, error) {
	query := `SELECT o.*, b.name AS bot_name FROM orders o
		JOIN bots b ON b.id = o.bot_id
		WHERE b.org_id = $1`
	args := []any{orgID}
	if source != "" {
		args = append(args, source)
		query += fmt.Sprintf(" AND o.source = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND o.status = $%d", len(args))
	}
	query += " ORDER BY o.created_at DESC"

	var orders []entity.Order
	if err := r.pg.DB().SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, fmt.Errorf("orders.ListByOrg: %w", err)
	}
	for i := range orders {
		if err := r.pg.DB().SelectContext(ctx, &orders[i].Items,
			"SELECT * FROM order_items WHERE order_id = $1 ORDER BY id",
			orders[i].ID); err != nil {
			return nil, fmt.Errorf("orders.ListByOrg items: %w", err)
		}
	}
	return orders, nil
}

func (r *Orders) UpdateStatus(ctx context.Context, orderID int, status string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"UPDATE orders SET status = $2 WHERE id = $1", orderID, status)
	if err != nil {
		return fmt.Errorf("orders.UpdateStatus: %w", err)
	}
	return nil
}

func (r *Orders) GetOrgID(ctx context.Context, orderID int) (int, error) {
	var orgID int
	err := r.pg.DB().GetContext(ctx, &orgID, `
		SELECT b.org_id FROM orders o
		JOIN bots b ON b.id = o.bot_id
		WHERE o.id = $1`, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("orders.GetOrgID: %w", err)
	}
	return orgID, nil
}
