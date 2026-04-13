package masterbot

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type handler struct {
	bot  *telego.Bot
	cfg  Config
	deps Deps

	// Legacy repos (kept for admin features)
	linksRepo      adminBotLinksRepo
	dashboardRepo  dashboardRepository
	campaignsRepo  campaignsRepository
	promotionsRepo promotionsRepository

	logger *slog.Logger
}

func newHandler(bot *telego.Bot, cfg Config, deps Deps, logger *slog.Logger) *handler {
	return &handler{
		bot:            bot,
		cfg:            cfg,
		deps:           deps,
		linksRepo:      deps.AdminLinks,
		dashboardRepo:  deps.Dashboard,
		campaignsRepo:  deps.Campaigns,
		promotionsRepo: deps.Promotions,
		logger:         logger,
	}
}

func (h *handler) Handle(ctx context.Context, update telego.Update) {
	// Handle managed bot updates (Bot API 9.6)
	if update.ManagedBot != nil {
		h.handleManagedBotUpdated(ctx, update)
		return
	}

	if update.Message == nil {
		return
	}

	msg := update.Message
	if msg.From == nil {
		return
	}

	text := strings.TrimSpace(msg.Text)

	switch {
	case text == "/start":
		h.handleStart(ctx, msg)
	case strings.HasPrefix(text, "/start "):
		// Deep link: /start {auth_token}
		token := strings.TrimPrefix(text, "/start ")
		h.handleStartWithToken(ctx, msg, token)
	case strings.HasPrefix(text, "/link "):
		h.handleLink(ctx, msg, strings.TrimPrefix(text, "/link "))
	case text == "/stats" || text == btnStats:
		h.handleStats(ctx, msg)
	case text == "/campaigns" || text == btnCampaigns:
		h.handleCampaigns(ctx, msg)
	case text == "/promotions" || text == btnPromos:
		h.handlePromotions(ctx, msg)
	case strings.HasPrefix(text, "/promo "):
		h.handleCreatePromo(ctx, msg, strings.TrimPrefix(text, "/promo "))
	case text == "/mybots":
		h.handleMyBots(ctx, msg)
	case text == "/help" || text == btnHelp:
		h.handleHelp(ctx, msg)
	default:
		// For linked users: unknown command help
		link := h.getLink(ctx, msg.From.ID)
		if link != nil {
			h.sendText(msg.Chat.ID, "Неизвестная команда. Отправьте /help для списка команд.")
		}
	}
}

// ── Deep link auth ──────────────────────────────────────────────────────────

func (h *handler) handleStartWithToken(ctx context.Context, msg *telego.Message, token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		h.handleStart(ctx, msg)
		return
	}

	if h.deps.AuthTokens == nil {
		h.handleStart(ctx, msg)
		return
	}

	// Validate and consume one-time token
	authData, err := h.deps.AuthTokens.ValidateAndConsume(ctx, token)
	if err != nil {
		h.logger.Warn("invalid auth token", "error", err, "telegram_id", msg.From.ID)
		h.sendText(msg.Chat.ID, "❌ Ссылка недействительна или истекла.\n\nПолучите новую ссылку на сайте.")
		return
	}

	// Create master bot link
	link := &entity.MasterBotLink{
		OrgID:            authData.OrgID,
		TelegramUserID:   msg.From.ID,
		TelegramUsername: msg.From.Username,
	}

	if err := h.deps.MasterLinks.CreateLink(ctx, link); err != nil {
		h.logger.Error("create master bot link", "error", err, "org_id", authData.OrgID)
		h.sendText(msg.Chat.ID, "Ошибка привязки. Попробуйте позже.")
		return
	}

	h.logger.Info("master bot linked",
		"org_id", authData.OrgID,
		"telegram_id", msg.From.ID,
		"username", msg.From.Username,
	)

	h.sendWithKeyboard(msg.Chat.ID,
		fmt.Sprintf("✅ Вы привязаны к организации #%d.\n\nВернитесь на сайт для настройки бота.", authData.OrgID),
		buildAdminMenu())
}

// ── Managed bot update ──────────────────────────────────────────────────────

func (h *handler) handleManagedBotUpdated(ctx context.Context, update telego.Update) {
	mbu := update.ManagedBot
	h.logger.Info("managed bot updated",
		"managed_bot_id", mbu.Bot.ID,
		"managed_bot_username", mbu.Bot.Username,
		"owner_id", mbu.User.ID,
	)

	if h.deps.Bots == nil {
		h.logger.Warn("bots repo not configured, cannot process managed bot update")
		return
	}

	// Get managed bot token
	tokenPtr, err := h.bot.GetManagedBotToken(ctx, &telego.GetManagedBotTokenParams{
		UserID: mbu.Bot.ID,
	})
	if err != nil {
		h.logger.Error("get managed bot token", "error", err, "bot_id", mbu.Bot.ID)
		return
	}
	if tokenPtr == nil {
		h.logger.Error("got nil token for managed bot", "bot_id", mbu.Bot.ID)
		return
	}
	botToken := *tokenPtr

	// Find pending bot by username match via owner's telegram_id
	ownerTgID := mbu.User.ID
	masterLink, err := h.deps.MasterLinks.GetLinkByTelegramID(ctx, ownerTgID)
	if err != nil || masterLink == nil {
		h.logger.Error("managed bot owner not linked", "owner_telegram_id", ownerTgID)
		return
	}

	// Find pending bot with matching username in this org
	bots, err := h.deps.Bots.GetByOrgID(ctx, masterLink.OrgID)
	if err != nil {
		h.logger.Error("get org bots", "error", err, "org_id", masterLink.OrgID)
		return
	}

	var pendingBot *entity.Bot
	for i := range bots {
		if bots[i].Status == "pending" && bots[i].Username == mbu.Bot.Username {
			pendingBot = &bots[i]
			break
		}
	}

	if pendingBot == nil {
		h.logger.Warn("no pending bot found for managed bot",
			"username", mbu.Bot.Username, "org_id", masterLink.OrgID)
		return
	}

	// Activate bot
	managedBotID := mbu.Bot.ID
	pendingBot.Token = botToken
	pendingBot.Status = "active"
	pendingBot.IsManagedBot = true
	pendingBot.ManagedBotID = &managedBotID
	pendingBot.CreatedByTelegramID = &ownerTgID

	if err := h.deps.Bots.Update(ctx, pendingBot); err != nil {
		h.logger.Error("activate managed bot", "error", err, "bot_id", pendingBot.ID)
		return
	}

	h.logger.Info("managed bot activated",
		"bot_id", pendingBot.ID,
		"username", mbu.Bot.Username,
		"org_id", masterLink.OrgID,
	)

	// Apply settings via Bot API
	h.applyBotSettings(ctx, pendingBot, botToken)

	// Notify owner
	h.sendText(ownerTgID, fmt.Sprintf("✅ Бот @%s создан и активирован!\n\nВернитесь на сайт для управления.", mbu.Bot.Username))
}

func (h *handler) applyBotSettings(ctx context.Context, bot *entity.Bot, token string) {
	tBot, err := telego.NewBot(token)
	if err != nil {
		h.logger.Error("create managed bot instance", "error", err)
		return
	}

	if bot.Settings.WelcomeMessage != "" {
		_ = tBot.SetMyDescription(ctx, &telego.SetMyDescriptionParams{
			Description: bot.Settings.WelcomeMessage,
		})
	}

	_ = tBot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
		Commands: []telego.BotCommand{
			{Command: "start", Description: "Начать"},
			{Command: "help", Description: "Помощь"},
		},
	})
}

// ── /mybots ─────────────────────────────────────────────────────────────────

func (h *handler) handleMyBots(ctx context.Context, msg *telego.Message) {
	if h.deps.MasterLinks == nil || h.deps.Bots == nil {
		h.sendText(msg.Chat.ID, "Функция пока недоступна.")
		return
	}

	masterLink, err := h.deps.MasterLinks.GetLinkByTelegramID(ctx, msg.From.ID)
	if err != nil || masterLink == nil {
		h.sendText(msg.Chat.ID, "⚠️ Ваш Telegram не привязан.\nАктивируйте бота через сайт.")
		return
	}

	bots, err := h.deps.Bots.GetByOrgID(ctx, masterLink.OrgID)
	if err != nil {
		h.sendText(msg.Chat.ID, "Ошибка получения списка ботов.")
		return
	}

	if len(bots) == 0 {
		h.sendText(msg.Chat.ID, "🤖 У вас пока нет ботов.\nСоздайте первого на сайте.")
		return
	}

	var sb strings.Builder
	sb.WriteString("🤖 Ваши боты:\n\n")
	for _, b := range bots {
		status := "⏳"
		switch b.Status {
		case "active":
			status = "✅"
		case "inactive":
			status = "⏸️"
		case "error":
			status = "❌"
		}
		username := b.Username
		if username != "" {
			username = "@" + username
		}
		sb.WriteString(fmt.Sprintf("%s %s %s\n", status, b.Name, username))
	}

	h.sendText(msg.Chat.ID, sb.String())
}

// ── Legacy auth (admin_bot_links) ───────────────────────────────────────────

func (h *handler) getLink(ctx context.Context, telegramID int64) *entity.AdminBotLink {
	if h.linksRepo == nil {
		return nil
	}
	link, err := h.linksRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil
	}
	return link
}

func (h *handler) requireAuth(ctx context.Context, msg *telego.Message) *entity.AdminBotLink {
	link := h.getLink(ctx, msg.From.ID)
	if link == nil {
		h.sendText(msg.Chat.ID, "⚠️ Ваш Telegram не привязан к аккаунту Revisitr.\n\n"+
			"Для привязки:\n"+
			"1. Откройте веб-панель → Настройки → Админ-бот\n"+
			"2. Нажмите «Получить код привязки»\n"+
			"3. Отправьте сюда: /link ВАШИ_КОД")
		return nil
	}
	return link
}

// ── Legacy commands ─────────────────────────────────────────────────────────

func (h *handler) handleStart(ctx context.Context, msg *telego.Message) {
	link := h.getLink(ctx, msg.From.ID)
	if link == nil {
		h.sendText(msg.Chat.ID,
			"👋 Добро пожаловать в Revisitr Bot!\n\n"+
				"Этот бот позволяет управлять заведением из Telegram:\n"+
				"• Создавать ботов для клиентов\n"+
				"• Готовить рассылки\n"+
				"• Смотреть статистику\n\n"+
				"Для привязки аккаунта активируйте бота через сайт\n"+
				"или отправьте: /link КОД")
		return
	}

	h.sendWithKeyboard(msg.Chat.ID,
		fmt.Sprintf("С возвращением! 👋\nОрганизация: #%d | Роль: %s", link.OrgID, link.Role),
		buildAdminMenu())
}

func (h *handler) handleLink(ctx context.Context, msg *telego.Message, code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		h.sendText(msg.Chat.ID, "Использование: /link КОД")
		return
	}

	existing := h.getLink(ctx, msg.From.ID)
	if existing != nil {
		h.sendText(msg.Chat.ID, "✅ Ваш аккаунт уже привязан.")
		return
	}

	link, err := h.linksRepo.GetByLinkCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			h.sendText(msg.Chat.ID, "❌ Неверный или истёкший код привязки.\nПолучите новый код в веб-панели.")
			return
		}
		h.logger.Error("get link by code", "error", err)
		h.sendText(msg.Chat.ID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if err := h.linksRepo.ActivateLink(ctx, link.ID, msg.From.ID); err != nil {
		h.logger.Error("activate link", "error", err, "link_id", link.ID)
		h.sendText(msg.Chat.ID, "Ошибка привязки. Попробуйте позже.")
		return
	}

	h.logger.Info("master bot linked (legacy)", "user_id", link.UserID, "telegram_id", msg.From.ID, "org_id", link.OrgID)
	h.sendWithKeyboard(msg.Chat.ID,
		"✅ Аккаунт успешно привязан!\n\nТеперь вы можете управлять заведением из Telegram.",
		buildAdminMenu())
}

func (h *handler) handleStats(ctx context.Context, msg *telego.Message) {
	link := h.requireAuth(ctx, msg)
	if link == nil {
		return
	}

	filter := entity.DashboardFilter{Period: "7d"}
	widgets, err := h.dashboardRepo.GetWidgets(ctx, link.OrgID, filter)
	if err != nil {
		h.logger.Error("get dashboard widgets", "error", err, "org_id", link.OrgID)
		h.sendText(msg.Chat.ID, "Ошибка получения статистики.")
		return
	}

	trendIcon := func(t float64) string {
		if t > 0 {
			return "📈"
		} else if t < 0 {
			return "📉"
		}
		return "➡️"
	}

	text := fmt.Sprintf("📊 Статистика за 7 дней\n\n"+
		"💰 Выручка: %.0f ₽ %s %.1f%%\n"+
		"🧾 Средний чек: %.0f ₽ %s %.1f%%\n"+
		"👤 Новых клиентов: %.0f %s %.1f%%\n"+
		"🔥 Активных клиентов: %.0f %s %.1f%%\n",
		widgets.Revenue.Value, trendIcon(widgets.Revenue.Trend), widgets.Revenue.Trend,
		widgets.AvgCheck.Value, trendIcon(widgets.AvgCheck.Trend), widgets.AvgCheck.Trend,
		widgets.NewClients.Value, trendIcon(widgets.NewClients.Trend), widgets.NewClients.Trend,
		widgets.ActiveClients.Value, trendIcon(widgets.ActiveClients.Trend), widgets.ActiveClients.Trend,
	)

	h.sendText(msg.Chat.ID, text)
}

func (h *handler) handleCampaigns(ctx context.Context, msg *telego.Message) {
	link := h.requireAuth(ctx, msg)
	if link == nil {
		return
	}

	campaigns, _, err := h.campaignsRepo.GetByOrgID(ctx, link.OrgID, 5, 0)
	if err != nil {
		h.logger.Error("get campaigns", "error", err, "org_id", link.OrgID)
		h.sendText(msg.Chat.ID, "Ошибка получения рассылок.")
		return
	}

	if len(campaigns) == 0 {
		h.sendText(msg.Chat.ID, "📬 Рассылок пока нет.\nСоздайте первую в веб-панели.")
		return
	}

	var sb strings.Builder
	sb.WriteString("📬 Последние рассылки:\n\n")
	for _, c := range campaigns {
		status := "⏳"
		switch c.Status {
		case "sent":
			status = "✅"
		case "sending":
			status = "🔄"
		case "scheduled":
			status = "📅"
		case "draft":
			status = "📝"
		case "failed":
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n   Статус: %s\n\n",
			status, c.Name, c.Status))
	}

	h.sendText(msg.Chat.ID, sb.String())
}

func (h *handler) handlePromotions(ctx context.Context, msg *telego.Message) {
	link := h.requireAuth(ctx, msg)
	if link == nil {
		return
	}

	promos, err := h.promotionsRepo.GetByOrgID(ctx, link.OrgID)
	if err != nil {
		h.logger.Error("get promotions", "error", err, "org_id", link.OrgID)
		h.sendText(msg.Chat.ID, "Ошибка получения акций.")
		return
	}

	if len(promos) == 0 {
		h.sendText(msg.Chat.ID, "🏷️ Акций пока нет.\nСоздайте первую в веб-панели.")
		return
	}

	var sb strings.Builder
	sb.WriteString("🏷️ Активные акции:\n\n")
	for _, p := range promos {
		if !p.Active {
			continue
		}
		sb.WriteString(fmt.Sprintf("• %s (%s)\n", p.Name, p.Type))
	}

	if sb.Len() == len("🏷️ Активные акции:\n\n") {
		sb.Reset()
		sb.WriteString("🏷️ Нет активных акций.")
	}

	h.sendText(msg.Chat.ID, sb.String())
}

func (h *handler) handleCreatePromo(ctx context.Context, msg *telego.Message, args string) {
	link := h.requireAuth(ctx, msg)
	if link == nil {
		return
	}

	if link.Role != "owner" {
		h.sendText(msg.Chat.ID, "⚠️ Только владелец может создавать промокоды.")
		return
	}

	code := strings.TrimSpace(args)
	if code == "" {
		b := make([]byte, 4)
		_, _ = rand.Read(b)
		code = strings.ToUpper(hex.EncodeToString(b))
	}

	pc := &entity.PromoCode{
		OrgID:  link.OrgID,
		Code:   code,
		Active: true,
	}

	if err := h.promotionsRepo.CreatePromoCode(ctx, pc); err != nil {
		h.logger.Error("create promo code", "error", err, "org_id", link.OrgID)
		h.sendText(msg.Chat.ID, "Ошибка создания промокода.")
		return
	}

	h.sendText(msg.Chat.ID, fmt.Sprintf("✅ Промокод создан: `%s`\n\nИспользуйте в рассылках или передайте клиентам.", code))
}

func (h *handler) handleHelp(_ context.Context, msg *telego.Message) {
	h.sendText(msg.Chat.ID,
		"📖 Команды Revisitr Bot:\n\n"+
			"/start — главное меню\n"+
			"/link КОД — привязать аккаунт\n"+
			"/mybots — список ваших ботов\n"+
			"/stats — статистика за 7 дней\n"+
			"/campaigns — последние рассылки\n"+
			"/promotions — активные акции\n"+
			"/promo [КОД] — создать промокод\n"+
			"/help — эта справка")
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func (h *handler) sendText(chatID int64, text string) {
	msg := tu.Message(tu.ID(chatID), text)
	msg = msg.WithParseMode("Markdown")
	if _, err := h.bot.SendMessage(context.Background(), msg); err != nil {
		h.logger.Error("send message", "error", err, "chat_id", chatID)
	}
}

func (h *handler) sendWithKeyboard(chatID int64, text string, kb *telego.ReplyKeyboardMarkup) {
	msg := tu.Message(tu.ID(chatID), text).WithReplyMarkup(kb)
	if _, err := h.bot.SendMessage(context.Background(), msg); err != nil {
		h.logger.Error("send message with keyboard", "error", err, "chat_id", chatID)
	}
}
