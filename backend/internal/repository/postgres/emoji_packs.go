package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type EmojiPacks struct {
	pg *Module
}

func NewEmojiPacks(pg *Module) *EmojiPacks {
	return &EmojiPacks{pg: pg}
}

func (r *EmojiPacks) Create(ctx context.Context, pack *entity.EmojiPack) error {
	query := `
		INSERT INTO emoji_packs (org_id, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.pg.DB().QueryRowContext(ctx, query,
		pack.OrgID, pack.Name, pack.SortOrder,
	).Scan(&pack.ID, &pack.CreatedAt, &pack.UpdatedAt)
}

func (r *EmojiPacks) GetByID(ctx context.Context, id int) (*entity.EmojiPack, error) {
	var pack entity.EmojiPack
	err := r.pg.DB().GetContext(ctx, &pack, "SELECT * FROM emoji_packs WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("emoji_packs.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("emoji_packs.GetByID: %w", err)
	}
	items, err := r.getItems(ctx, id)
	if err != nil {
		return nil, err
	}
	pack.Items = items
	return &pack, nil
}

func (r *EmojiPacks) GetByOrgID(ctx context.Context, orgID int) ([]entity.EmojiPack, error) {
	var packs []entity.EmojiPack
	err := r.pg.DB().SelectContext(ctx, &packs,
		"SELECT * FROM emoji_packs WHERE org_id = $1 ORDER BY sort_order ASC, created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("emoji_packs.GetByOrgID: %w", err)
	}
	for i := range packs {
		items, err := r.getItems(ctx, packs[i].ID)
		if err != nil {
			return nil, err
		}
		packs[i].Items = items
	}
	return packs, nil
}

func (r *EmojiPacks) Update(ctx context.Context, pack *entity.EmojiPack) error {
	query := `
		UPDATE emoji_packs
		SET name = $1, sort_order = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at`
	err := r.pg.DB().QueryRowContext(ctx, query,
		pack.Name, pack.SortOrder, pack.ID,
	).Scan(&pack.UpdatedAt)
	if err != nil {
		return fmt.Errorf("emoji_packs.Update: %w", err)
	}
	return nil
}

func (r *EmojiPacks) Delete(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM emoji_packs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("emoji_packs.Delete: %w", err)
	}
	return nil
}

func (r *EmojiPacks) CreateItem(ctx context.Context, item *entity.EmojiItem) error {
	query := `
		INSERT INTO emoji_items (pack_id, name, image_url, sort_order, tg_sticker_set, tg_custom_emoji_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return r.pg.DB().QueryRowContext(ctx, query,
		item.PackID, item.Name, item.ImageURL, item.SortOrder, item.TgStickerSet, item.TgCustomEmojiID,
	).Scan(&item.ID, &item.CreatedAt)
}

func (r *EmojiPacks) UpdateItem(ctx context.Context, item *entity.EmojiItem) error {
	query := `
		UPDATE emoji_items
		SET name = $1, sort_order = $2
		WHERE id = $3`
	_, err := r.pg.DB().ExecContext(ctx, query, item.Name, item.SortOrder, item.ID)
	if err != nil {
		return fmt.Errorf("emoji_packs.UpdateItem: %w", err)
	}
	return nil
}

func (r *EmojiPacks) DeleteItem(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM emoji_items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("emoji_packs.DeleteItem: %w", err)
	}
	return nil
}

func (r *EmojiPacks) GetItemByID(ctx context.Context, id int) (*entity.EmojiItem, error) {
	var item entity.EmojiItem
	err := r.pg.DB().GetContext(ctx, &item, "SELECT * FROM emoji_items WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("emoji_packs.GetItemByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("emoji_packs.GetItemByID: %w", err)
	}
	return &item, nil
}

func (r *EmojiPacks) ReorderItems(ctx context.Context, packID int, itemIDs []int) error {
	for i, id := range itemIDs {
		_, err := r.pg.DB().ExecContext(ctx,
			"UPDATE emoji_items SET sort_order = $1 WHERE id = $2 AND pack_id = $3",
			i, id, packID)
		if err != nil {
			return fmt.Errorf("emoji_packs.ReorderItems: %w", err)
		}
	}
	return nil
}

func (r *EmojiPacks) getItems(ctx context.Context, packID int) ([]entity.EmojiItem, error) {
	var items []entity.EmojiItem
	err := r.pg.DB().SelectContext(ctx, &items,
		"SELECT * FROM emoji_items WHERE pack_id = $1 ORDER BY sort_order ASC", packID)
	if err != nil {
		return nil, fmt.Errorf("emoji_packs.getItems: %w", err)
	}
	return items, nil
}
