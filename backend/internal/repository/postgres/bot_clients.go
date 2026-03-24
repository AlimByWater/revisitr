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
		INSERT INTO bot_clients (bot_id, telegram_id, username, first_name, last_name, phone, phone_normalized, qr_code)
		VALUES (:bot_id, :telegram_id, :username, :first_name, :last_name, :phone, :phone_normalized, :qr_code)
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

// UpdateRFM sets RFM scores and segment label for a single bot client.
func (r *BotClients) UpdateRFM(ctx context.Context, clientID, recency, frequency int, monetary float64, segment string) error {
	query := `
		UPDATE bot_clients
		SET rfm_recency = $1, rfm_frequency = $2, rfm_monetary = $3,
		    rfm_segment = $4, rfm_updated_at = NOW()
		WHERE id = $5`

	result, err := r.pg.DB().ExecContext(ctx, query, recency, frequency, monetary, segment, clientID)
	if err != nil {
		return fmt.Errorf("bot_clients.UpdateRFM: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bot_clients.UpdateRFM rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bot_clients.UpdateRFM: %w", sql.ErrNoRows)
	}
	return nil
}

// GetAllByOrgID returns all bot clients for the given org (used by RFM service).
func (r *BotClients) GetAllByOrgID(ctx context.Context, orgID int) ([]entity.BotClient, error) {
	var clients []entity.BotClient
	query := `
		SELECT bc.*
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1`
	if err := r.pg.DB().SelectContext(ctx, &clients, query, orgID); err != nil {
		return nil, fmt.Errorf("bot_clients.GetAllByOrgID: %w", err)
	}
	return clients, nil
}

// GetByPhone finds a bot client by phone number within the given org.
// Uses phone_normalized for reliable matching, falls back to raw phone.
func (r *BotClients) GetByPhone(ctx context.Context, orgID int, phone string) (*entity.BotClient, error) {
	normalized := entity.NormalizePhone(phone)
	if normalized == "" {
		normalized = phone
	}

	var client entity.BotClient
	query := `
		SELECT bc.*
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1 AND (bc.phone_normalized = $2 OR bc.phone = $3)
		LIMIT 1`
	err := r.pg.DB().GetContext(ctx, &client, query, orgID, normalized, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bot_clients.GetByPhone: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("bot_clients.GetByPhone: %w", err)
	}
	return &client, nil
}

// GetByQRCode finds a bot client by their unique QR code identifier.
func (r *BotClients) GetByQRCode(ctx context.Context, qrCode string) (*entity.BotClient, error) {
	var client entity.BotClient
	err := r.pg.DB().GetContext(ctx, &client,
		"SELECT * FROM bot_clients WHERE qr_code = $1", qrCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bot_clients.GetByQRCode: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("bot_clients.GetByQRCode: %w", err)
	}
	return &client, nil
}
