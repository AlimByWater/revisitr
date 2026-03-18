package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type BotClients struct {
	pg *Module
}

func NewBotClients(pg *Module) *BotClients {
	return &BotClients{pg: pg}
}

func (r *BotClients) Create(ctx context.Context, client *entity.BotClient) error {
	query := `
		INSERT INTO bot_clients (bot_id, telegram_id, username, first_name, last_name, phone)
		VALUES (:bot_id, :telegram_id, :username, :first_name, :last_name, :phone)
		RETURNING id, registered_at`

	rows, err := r.pg.DB().NamedQueryContext(ctx, query, client)
	if err != nil {
		return fmt.Errorf("botClients.Create: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&client.ID, &client.RegisteredAt); err != nil {
			return fmt.Errorf("botClients.Create scan: %w", err)
		}
	}

	return nil
}

func (r *BotClients) GetByTelegramID(ctx context.Context, botID int, telegramID int64) (*entity.BotClient, error) {
	var client entity.BotClient
	err := r.pg.DB().GetContext(ctx, &client,
		"SELECT * FROM bot_clients WHERE bot_id = $1 AND telegram_id = $2", botID, telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("botClients.GetByTelegramID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("botClients.GetByTelegramID: %w", err)
	}
	return &client, nil
}

func (r *BotClients) GetByBotID(ctx context.Context, botID int, limit, offset int) ([]entity.BotClient, int, error) {
	var total int
	err := r.pg.DB().GetContext(ctx, &total,
		"SELECT COUNT(*) FROM bot_clients WHERE bot_id = $1", botID)
	if err != nil {
		return nil, 0, fmt.Errorf("botClients.GetByBotID count: %w", err)
	}

	var clients []entity.BotClient
	err = r.pg.DB().SelectContext(ctx, &clients,
		"SELECT * FROM bot_clients WHERE bot_id = $1 ORDER BY registered_at DESC LIMIT $2 OFFSET $3",
		botID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("botClients.GetByBotID: %w", err)
	}

	return clients, total, nil
}

func (r *BotClients) CountByBotID(ctx context.Context, botID int) (int, error) {
	var count int
	err := r.pg.DB().GetContext(ctx, &count,
		"SELECT COUNT(*) FROM bot_clients WHERE bot_id = $1", botID)
	if err != nil {
		return 0, fmt.Errorf("botClients.CountByBotID: %w", err)
	}
	return count, nil
}
