package botmanager

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"revisitr/internal/entity"
	"revisitr/internal/service/eventbus"
	"revisitr/internal/service/poscode"
	tgService "revisitr/internal/service/telegram"
	walletUC "revisitr/internal/usecase/wallet"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type botsRepository interface {
	GetAllActive(ctx context.Context) ([]entity.Bot, error)
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
	UpdateSettings(ctx context.Context, id int, settings entity.BotSettings) error
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

type menusRepository interface {
	GetBotPOSLocations(ctx context.Context, botID int) ([]int, error)
	GetActiveMenuForPOS(ctx context.Context, orgID, posID int) (*entity.Menu, error)
}

type moduleSettingsRepository interface {
	Get(ctx context.Context, botID int, moduleKey string) (*entity.BotModuleSettings, error)
}

type lunchRepository interface {
	GetFullProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error)
	CreateOrder(ctx context.Context, order *entity.LunchOrder) error
}

type lunchEventPublisher interface {
	PublishLunchOrderCreated(ctx context.Context, event eventbus.LunchOrderEvent) error
}

type emojiRepository interface {
	GetSyncedItemsByOrgID(ctx context.Context, orgID int) ([]entity.EmojiItem, error)
}

type sessionStore interface {
	Load(ctx context.Context, botID int, chatID int64) (*FlowState, error)
	Save(ctx context.Context, botID int, chatID int64, state FlowState) error
	Delete(ctx context.Context, botID int, chatID int64) error
}

type walletUsecase interface {
	GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error)
	IssuePass(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error)
	GetClientPasses(ctx context.Context, clientID int) ([]entity.WalletPass, error)
	RefreshPassBalance(ctx context.Context, clientID int, balance int, level string) error
	GetClientsQRCode(ctx context.Context, clientID int) (string, error)
	GetOrgName(ctx context.Context, orgID int) (string, error)
	GenerateGoogleSaveURL(ctx context.Context, orgID int, pass *entity.WalletPass) (string, error)
}

type botInstance struct {
	bot    *telego.Bot
	cancel context.CancelFunc
	info   entity.Bot
}

type Manager struct {
	botsRepo           botsRepository
	clientsRepo        botClientsRepository
	loyaltyRepo        loyaltyRepository
	posRepo            posRepository
	menusRepo          menusRepository
	lunchRepo          lunchRepository
	lunchEvents        lunchEventPublisher
	emojiRepo          emojiRepository
	moduleSettingsRepo moduleSettingsRepository
	sessions           sessionStore
	wallet             walletUsecase
	passGen            *walletUC.PassGenerator
	posCode            *poscode.Service
	baseURL            string
	tgSender           *tgService.Sender
	logger             *slog.Logger
	apiServer          string // custom Telegram Bot API server URL (empty = default)
	proxyURL           string // HTTP proxy URL for Telegram API (empty = direct)
	adminBotToken      string
	adminBot           *telego.Bot

	mu        sync.RWMutex
	instances map[int]*botInstance
}

type ManagerOption func(*Manager)

func WithTelegramSender(ts *tgService.Sender) ManagerOption {
	return func(m *Manager) { m.tgSender = ts }
}

func WithAPIServer(url string) ManagerOption {
	return func(m *Manager) { m.apiServer = url }
}

func WithProxy(url string) ManagerOption {
	return func(m *Manager) { m.proxyURL = url }
}

func WithMenus(repo menusRepository) ManagerOption {
	return func(m *Manager) { m.menusRepo = repo }
}

func WithEmoji(repo emojiRepository) ManagerOption {
	return func(m *Manager) { m.emojiRepo = repo }
}

func WithLunch(repo lunchRepository) ManagerOption {
	return func(m *Manager) { m.lunchRepo = repo }
}

func WithLunchEvents(pub lunchEventPublisher) ManagerOption {
	return func(m *Manager) { m.lunchEvents = pub }
}

func WithSessionStore(store sessionStore) ManagerOption {
	return func(m *Manager) { m.sessions = store }
}

func WithModuleSettings(repo moduleSettingsRepository) ManagerOption {
	return func(m *Manager) { m.moduleSettingsRepo = repo }
}

func WithPOSCode(svc *poscode.Service) ManagerOption {
	return func(m *Manager) { m.posCode = svc }
}

func WithAdminBotToken(token string) ManagerOption {
	return func(m *Manager) { m.adminBotToken = token }
}

func WithWallet(uc walletUsecase) ManagerOption {
	return func(m *Manager) { m.wallet = uc }
}

func WithBaseURL(url string) ManagerOption {
	return func(m *Manager) { m.baseURL = url }
}

// botOpts returns telego options configured with API server and proxy.
func (m *Manager) botOpts() []telego.BotOption {
	var opts []telego.BotOption
	if m.apiServer != "" {
		opts = append(opts, telego.WithAPIServer(m.apiServer))
	}
	if m.proxyURL != "" {
		proxyParsed, err := url.Parse(m.proxyURL)
		if err == nil {
			opts = append(opts, telego.WithHTTPClient(&http.Client{
				Transport: &http.Transport{Proxy: http.ProxyURL(proxyParsed)},
			}))
		}
	}
	return opts
}

func New(
	botsRepo botsRepository,
	clientsRepo botClientsRepository,
	loyaltyRepo loyaltyRepository,
	posRepo posRepository,
	logger *slog.Logger,
	opts ...ManagerOption,
) *Manager {
	m := &Manager{
		botsRepo:    botsRepo,
		clientsRepo: clientsRepo,
		loyaltyRepo: loyaltyRepo,
		posRepo:     posRepo,
		logger:      logger,
		passGen:     walletUC.NewPassGenerator(),
		instances:   make(map[int]*botInstance),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Start(ctx context.Context) error {
	if m.adminBot == nil && m.adminBotToken != "" {
		adminBot, err := telego.NewBot(m.adminBotToken, m.botOpts()...)
		if err != nil {
			return fmt.Errorf("botmanager: create admin bot: %w", err)
		}
		m.adminBot = adminBot
	}

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
		m.logger.Info("bot stopped", "bot_id", id, "name", inst.info.Name)
	}

	m.instances = make(map[int]*botInstance)
	m.logger.Info("all bots stopped")
}

func (m *Manager) startBot(parentCtx context.Context, b entity.Bot) error {
	tBot, err := telego.NewBot(b.Token, m.botOpts()...)
	if err != nil {
		return fmt.Errorf("create telego bot %q: %w", b.Name, err)
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

	handler := newHandler(m, tBot, &inst.info)

	go func() {
		updates, err := tBot.UpdatesViaLongPolling(ctx, nil)
		if err != nil {
			m.logger.Error("long polling failed", "bot_id", b.ID, "error", err)
			return
		}

		m.logger.Info("bot started", "bot_id", b.ID, "name", b.Name, "username", b.Username)

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

// ── BotEventHandler implementation ──────────────────────────────────────────

func (m *Manager) OnBotReload(ctx context.Context, botID int) error {
	m.logger.Info("event: bot reload", "bot_id", botID)
	return m.ReloadBot(ctx, botID)
}

func (m *Manager) OnBotStop(ctx context.Context, botID int) error {
	m.logger.Info("event: bot stop", "bot_id", botID)
	return m.RemoveBot(botID)
}

func (m *Manager) OnBotStart(ctx context.Context, botID int) error {
	m.logger.Info("event: bot start", "bot_id", botID)
	return m.AddBot(ctx, botID)
}

// OnNotifyClient sends a plain-text message to a client's chat via the running
// bot instance. Best-effort: no-ops if the bot is not running and never fails
// the caller (the originating POS operation already succeeded).
func (m *Manager) OnNotifyClient(_ context.Context, botID int, chatID int64, text string) error {
	if text == "" {
		return nil
	}

	m.mu.RLock()
	inst, ok := m.instances[botID]
	m.mu.RUnlock()
	if !ok {
		m.logger.Warn("notify client: bot not running", "bot_id", botID, "chat_id", chatID)
		return nil
	}

	if _, err := inst.bot.SendMessage(context.Background(), tu.Message(tu.ID(chatID), text)); err != nil {
		m.logger.Error("notify client: send failed", "bot_id", botID, "chat_id", chatID, "error", err)
	}
	return nil
}

// OnBotSettingsChanged performs a hot update of bot settings without restarting long polling.
func (m *Manager) OnBotSettingsChanged(ctx context.Context, botID int, field string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[botID]
	if !ok {
		return nil // bot not running, nothing to update
	}

	bot, err := m.botsRepo.GetByID(ctx, botID)
	if err != nil {
		return fmt.Errorf("botmanager: hot update bot %d: %w", botID, err)
	}

	inst.info = *bot
	m.logger.Info("event: bot settings updated",
		"bot_id", botID,
		"field", field,
	)
	return nil
}
