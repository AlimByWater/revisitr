package adminbot

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
	bot            *telego.Bot
	linksRepo      adminBotLinksRepo
	dashboardRepo  dashboardRepository
	campaignsRepo  campaignsRepository
	promotionsRepo promotionsRepository
	logger         *slog.Logger
}

func newHandler(
	bot *telego.Bot,
	linksRepo adminBotLinksRepo,
	dashboardRepo dashboardRepository,
	campaignsRepo campaignsRepository,
	promotionsRepo promotionsRepository,
	_ interface{}, // analyticsRepo placeholder for future use
	logger *slog.Logger,
) *handler {
	return &handler{
		bot:            bot,
		linksRepo:      linksRepo,
		dashboardRepo:  dashboardRepo,
		campaignsRepo:  campaignsRepo,
		promotionsRepo: promotionsRepo,
		logger:         logger,
	}
}

func (h *handler) Handle(ctx context.Context, update telego.Update) {
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
	case text == "/help" || text == btnHelp:
		h.handleHelp(ctx, msg)
	default:
		// Ignore unknown messages for non-linked users; show help for linked
		link := h.getLink(ctx, msg.From.ID)
		if link != nil {
			h.sendText(msg.Chat.ID, "Неизвестная команда. Отправьте /help для списка команд.")
		}
	}
}

// ── Auth ──────────────────────────────────────────────────────────────────────

func (h *handler) getLink(ctx context.Context, telegramID int64) *entity.AdminBotLink {
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

// ── Commands ──────────────────────────────────────────────────────────────────

func (h *handler) handleStart(ctx context.Context, msg *telego.Message) {
	link := h.getLink(ctx, msg.From.ID)
	if link == nil {
		h.sendText(msg.Chat.ID,
			"👋 Добро пожаловать в Revisitr Admin Bot!\n\n"+
				"Этот бот позволяет управлять вашим заведением прямо из Telegram:\n"+
				"• Смотреть статистику\n"+
				"• Управлять рассылками и промокодами\n"+
				"• Получать уведомления\n\n"+
				"Для начала привяжите аккаунт:\n"+
				"/link КОД — привязать аккаунт по коду из веб-панели")
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

	// Check if already linked
	existing := h.getLink(ctx, msg.From.ID)
	if existing != nil {
		h.sendText(msg.Chat.ID, "✅ Ваш аккаунт уже привязан.")
		return
	}

	// Find link by code
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

	// Activate the link
	if err := h.linksRepo.ActivateLink(ctx, link.ID, msg.From.ID); err != nil {
		h.logger.Error("activate link", "error", err, "link_id", link.ID)
		h.sendText(msg.Chat.ID, "Ошибка привязки. Попробуйте позже.")
		return
	}

	h.logger.Info("admin bot linked", "user_id", link.UserID, "telegram_id", msg.From.ID, "org_id", link.OrgID)
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

	// Generate a random code if not specified
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
		"📖 Команды Admin Bot:\n\n"+
			"/start — главное меню\n"+
			"/link КОД — привязать аккаунт\n"+
			"/stats — статистика за 7 дней\n"+
			"/campaigns — последние рассылки\n"+
			"/promotions — активные акции\n"+
			"/promo [КОД] — создать промокод\n"+
			"/help — эта справка")
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (h *handler) sendText(chatID int64, text string) {
	msg := tu.Message(tu.ID(chatID), text)
	msg = msg.WithParseMode("Markdown")
	if _, err := h.bot.SendMessage(msg); err != nil {
		h.logger.Error("send message", "error", err, "chat_id", chatID)
	}
}

func (h *handler) sendWithKeyboard(chatID int64, text string, kb *telego.ReplyKeyboardMarkup) {
	msg := tu.Message(tu.ID(chatID), text).WithReplyMarkup(kb)
	if _, err := h.bot.SendMessage(msg); err != nil {
		h.logger.Error("send message with keyboard", "error", err, "chat_id", chatID)
	}
}
