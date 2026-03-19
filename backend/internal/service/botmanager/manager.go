package botmanager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
)

type botsRepository interface {
	GetAllActive(ctx context.Context) ([]entity.Bot, error)
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type botClientsRepository interface {
	Create(ctx context.Context, client *entity.BotClient) error
	GetByTelegramID(ctx context.Context, botID int, telegramID int64) (*entity.BotClient, error)
}

type loyaltyRepository interface {
	GetProgramsByOrgID(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	GetClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	GetLevelsByProgramID(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error)
	UpsertClientLoyalty(ctx context.Context, cl *entity.ClientLoyalty) error
	CreateTransaction(ctx context.Context, tx *entity.LoyaltyTransaction) error
}

type posRepository interface {
	GetByOrgID(ctx context.Context, orgID int) ([]entity.POSLocation, error)
}

type botInstance struct {
	bot    *telego.Bot
	cancel context.CancelFunc
	info   entity.Bot
}

type Manager struct {
	botsRepo    botsRepository
	clientsRepo botClientsRepository
	loyaltyRepo loyaltyRepository
	posRepo     posRepository
	logger      *slog.Logger

	mu        sync.RWMutex
	instances map[int]*botInstance
}

func New(
	botsRepo botsRepository,
	clientsRepo botClientsRepository,
	loyaltyRepo loyaltyRepository,
	posRepo posRepository,
	logger *slog.Logger,
) *Manager {
	return &Manager{
		botsRepo:    botsRepo,
		clientsRepo: clientsRepo,
		loyaltyRepo: loyaltyRepo,
		posRepo:     posRepo,
		logger:      logger,
		instances:   make(map[int]*botInstance),
	}
}

func (m *Manager) Start(ctx context.Context) error {
	bots, err := m.botsRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("botmanager: load active bots: %w", err)
	}

	m.logger.Info("loading active bots", "count", len(bots))

	for _, b := range bots {
		if err := m.startBot(ctx, b); err != nil {
			m.logger.Error("failed to start bot", "bot_id", b.ID, "name", b.Name, "error", err)
			continue
		}
	}

	return nil
}

func (m *Manager) AddBot(ctx context.Context, botID int) error {
	b, err := m.botsRepo.GetByID(ctx, botID)
	if err != nil {
		return fmt.Errorf("botmanager: get bot %d: %w", botID, err)
	}

	m.mu.RLock()
	_, exists := m.instances[botID]
	m.mu.RUnlock()

	if exists {
		return m.ReloadBot(ctx, botID)
	}

	return m.startBot(ctx, *b)
}

func (m *Manager) RemoveBot(botID int) error {
	m.mu.Lock()
	inst, exists := m.instances[botID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	delete(m.instances, botID)
	m.mu.Unlock()

	inst.cancel()
	inst.bot.StopLongPolling()
	m.logger.Info("bot stopped", "bot_id", botID, "name", inst.info.Name)
	return nil
}

func (m *Manager) ReloadBot(ctx context.Context, botID int) error {
	if err := m.RemoveBot(botID); err != nil {
		return err
	}

	b, err := m.botsRepo.GetByID(ctx, botID)
	if err != nil {
		return fmt.Errorf("botmanager: reload bot %d: %w", botID, err)
	}

	if b.Status != "active" {
		m.logger.Info("bot not active, skipping reload", "bot_id", botID)
		return nil
	}

	return m.startBot(ctx, *b)
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, inst := range m.instances {
		inst.cancel()
		inst.bot.StopLongPolling()
		m.logger.Info("bot stopped", "bot_id", id, "name", inst.info.Name)
	}

	m.instances = make(map[int]*botInstance)
	m.logger.Info("all bots stopped")
}

func (m *Manager) startBot(parentCtx context.Context, b entity.Bot) error {
	tBot, err := telego.NewBot(b.Token)
	if err != nil {
		return fmt.Errorf("create telego bot %q: %w", b.Name, err)
	}

	// Verify token by getting bot info
	info, err := tBot.GetMe()
	if err != nil {
		return fmt.Errorf("verify bot token %q: %w", b.Name, err)
	}

	ctx, cancel := context.WithCancel(parentCtx)

	inst := &botInstance{
		bot:    tBot,
		cancel: cancel,
		info:   b,
	}

	m.mu.Lock()
	m.instances[b.ID] = inst
	m.mu.Unlock()

	handler := newHandler(m, tBot, b)

	go func() {
		updates, err := tBot.UpdatesViaLongPolling(nil)
		if err != nil {
			m.logger.Error("long polling failed", "bot_id", b.ID, "error", err)
			return
		}

		m.logger.Info("bot started", "bot_id", b.ID, "name", b.Name, "username", info.Username)

		for {
			select {
			case <-ctx.Done():
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				handler.Handle(ctx, update)
			}
		}
	}()

	return nil
}

func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.instances)
}
