package botmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type handler struct {
	mgr    *Manager
	bot    *telego.Bot
	info   entity.Bot
	logger *slog.Logger
}

func newHandler(mgr *Manager, bot *telego.Bot, info entity.Bot) *handler {
	return &handler{
		mgr:    mgr,
		bot:    bot,
		info:   info,
		logger: mgr.logger.With("bot_id", info.ID, "bot_name", info.Name),
	}
}

func (h *handler) Handle(ctx context.Context, update telego.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID

	// Handle contact sharing (registration step)
	if msg.Contact != nil {
		h.handleContact(ctx, msg)
		return
	}

	text := strings.TrimSpace(msg.Text)

	switch {
	case text == "/start":
		h.handleStart(ctx, msg)
	case text == btnBalance:
		h.handleBalance(ctx, msg)
	case text == btnLocations:
		h.handleLocations(ctx, msg)
	case text == btnAbout:
		h.handleAbout(ctx, msg)
	case text == btnBack:
		h.sendMainMenu(ctx, chatID, "Главное меню")
	default:
		// Check if it matches a custom button
		h.handleCustomButton(ctx, msg, text)
	}
}

func (h *handler) handleStart(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	telegramID := msg.From.ID

	// Check if user already registered
	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, telegramID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		h.logger.Error("check client registration", "error", err, "telegram_id", telegramID)
		h.sendText(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if client != nil {
		// Already registered — show main menu
		welcome := h.info.Settings.WelcomeMessage
		if welcome == "" {
			welcome = fmt.Sprintf("С возвращением, %s! 👋", client.FirstName)
		}
		h.sendMainMenu(ctx, chatID, welcome)
		return
	}

	// New user — start registration
	welcome := h.info.Settings.WelcomeMessage
	if welcome == "" {
		welcome = fmt.Sprintf("Добро пожаловать в %s! 🎉", h.info.Name)
	}

	// If registration form requires phone, ask for it
	needsPhone := false
	for _, field := range h.info.Settings.RegistrationForm {
		if field.Name == "phone" && field.Required {
			needsPhone = true
			break
		}
	}

	if needsPhone {
		h.sendWithKeyboard(chatID, welcome+"\n\nДля регистрации поделитесь номером телефона:", buildContactRequest())
	} else {
		// Auto-register from Telegram profile
		h.autoRegister(ctx, msg)
	}
}

func (h *handler) handleContact(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	telegramID := msg.From.ID
	contact := msg.Contact

	// Check not already registered
	existing, _ := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, telegramID)
	if existing != nil {
		h.sendMainMenu(ctx, chatID, "Вы уже зарегистрированы! 👍")
		return
	}

	client := &entity.BotClient{
		BotID:      h.info.ID,
		TelegramID: telegramID,
		Username:   msg.From.Username,
		FirstName:  msg.From.FirstName,
		LastName:   msg.From.LastName,
		Phone:      contact.PhoneNumber,
	}

	if err := h.mgr.clientsRepo.Create(ctx, client); err != nil {
		h.logger.Error("create client", "error", err, "telegram_id", telegramID)
		h.sendText(chatID, "Ошибка регистрации. Попробуйте позже.")
		return
	}

	h.logger.Info("client registered", "client_id", client.ID, "telegram_id", telegramID)

	// Award welcome bonus if loyalty program exists
	h.awardWelcomeBonus(ctx, client)

	h.sendMainMenu(ctx, chatID, fmt.Sprintf("Добро пожаловать, %s! Вы успешно зарегистрированы. 🎉", client.FirstName))
}

func (h *handler) autoRegister(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	telegramID := msg.From.ID

	client := &entity.BotClient{
		BotID:      h.info.ID,
		TelegramID: telegramID,
		Username:   msg.From.Username,
		FirstName:  msg.From.FirstName,
		LastName:   msg.From.LastName,
	}

	if err := h.mgr.clientsRepo.Create(ctx, client); err != nil {
		h.logger.Error("auto-register client", "error", err, "telegram_id", telegramID)
		h.sendText(chatID, "Ошибка регистрации. Попробуйте позже.")
		return
	}

	h.logger.Info("client auto-registered", "client_id", client.ID, "telegram_id", telegramID)

	h.awardWelcomeBonus(ctx, client)

	h.sendMainMenu(ctx, chatID, fmt.Sprintf("Добро пожаловать, %s! Вы успешно зарегистрированы. 🎉", client.FirstName))
}

func (h *handler) handleBalance(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	telegramID := msg.From.ID

	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, telegramID)
	if err != nil {
		h.sendText(chatID, "Сначала нужно зарегистрироваться. Отправьте /start")
		return
	}

	programs, err := h.mgr.loyaltyRepo.GetProgramsByOrgID(ctx, h.info.OrgID)
	if err != nil || len(programs) == 0 {
		h.sendText(chatID, "Программа лояльности пока не настроена.")
		return
	}

	var sb strings.Builder
	sb.WriteString("💰 Ваш баланс:\n\n")

	for _, prog := range programs {
		if !prog.IsActive {
			continue
		}

		cl, err := h.mgr.loyaltyRepo.GetClientLoyalty(ctx, client.ID, prog.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			continue
		}

		balance := 0.0
		levelName := "—"

		if cl != nil {
			balance = cl.Balance
			if cl.LevelID != nil {
				levels, _ := h.mgr.loyaltyRepo.GetLevelsByProgramID(ctx, prog.ID)
				for _, l := range levels {
					if l.ID == *cl.LevelID {
						levelName = l.Name
						break
					}
				}
			}
		}

		currencyName := prog.Config.CurrencyName
		if currencyName == "" {
			currencyName = "баллов"
		}

		sb.WriteString(fmt.Sprintf("📋 %s\n", prog.Name))
		sb.WriteString(fmt.Sprintf("   Баланс: %.0f %s\n", balance, currencyName))
		sb.WriteString(fmt.Sprintf("   Уровень: %s\n\n", levelName))
	}

	h.sendText(chatID, sb.String())
}

func (h *handler) handleLocations(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID

	locations, err := h.mgr.posRepo.GetByOrgID(ctx, h.info.OrgID)
	if err != nil || len(locations) == 0 {
		h.sendText(chatID, "Информация о точках пока недоступна.")
		return
	}

	var sb strings.Builder
	sb.WriteString("📍 Наши точки:\n\n")

	for _, loc := range locations {
		if !loc.IsActive {
			continue
		}
		sb.WriteString(fmt.Sprintf("🏠 %s\n", loc.Name))
		if loc.Address != "" {
			sb.WriteString(fmt.Sprintf("   📌 %s\n", loc.Address))
		}
		if loc.Phone != "" {
			sb.WriteString(fmt.Sprintf("   📞 %s\n", loc.Phone))
		}
		sb.WriteString("\n")
	}

	h.sendText(chatID, sb.String())
}

func (h *handler) handleAbout(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	h.sendText(chatID, fmt.Sprintf("ℹ️ %s\n\nИспользуйте меню для навигации.", h.info.Name))
}

func (h *handler) handleCustomButton(ctx context.Context, msg *telego.Message, text string) {
	for _, btn := range h.info.Settings.Buttons {
		if btn.Label == text {
			switch btn.Type {
			case "url":
				h.sendText(msg.Chat.ID, btn.Value)
			case "text":
				h.sendText(msg.Chat.ID, btn.Value)
			default:
				h.sendText(msg.Chat.ID, btn.Value)
			}
			return
		}
	}
}

func (h *handler) awardWelcomeBonus(ctx context.Context, client *entity.BotClient) {
	programs, err := h.mgr.loyaltyRepo.GetProgramsByOrgID(ctx, h.info.OrgID)
	if err != nil {
		return
	}

	for _, prog := range programs {
		if !prog.IsActive || prog.Config.WelcomeBonus <= 0 {
			continue
		}

		bonus := float64(prog.Config.WelcomeBonus)
		cl := &entity.ClientLoyalty{
			ClientID:    client.ID,
			ProgramID:   prog.ID,
			Balance:     bonus,
			TotalEarned: bonus,
		}

		if err := h.mgr.loyaltyRepo.UpsertClientLoyalty(ctx, cl); err != nil {
			h.logger.Error("award welcome bonus", "error", err, "client_id", client.ID, "program_id", prog.ID)
			continue
		}

		tx := &entity.LoyaltyTransaction{
			ClientID:     client.ID,
			ProgramID:    prog.ID,
			Type:         "earn",
			Amount:       bonus,
			BalanceAfter: bonus,
			Description:  "Приветственный бонус",
		}
		if err := h.mgr.loyaltyRepo.CreateTransaction(ctx, tx); err != nil {
			h.logger.Error("create welcome bonus tx", "error", err)
		}

		currencyName := prog.Config.CurrencyName
		if currencyName == "" {
			currencyName = "баллов"
		}

		h.sendText(client.TelegramID, fmt.Sprintf("🎁 Вам начислен приветственный бонус: %.0f %s!", bonus, currencyName))
		h.logger.Info("welcome bonus awarded", "client_id", client.ID, "amount", bonus, "program", prog.Name)
	}
}

func (h *handler) sendMainMenu(_ context.Context, chatID int64, text string) {
	h.sendWithKeyboard(chatID, text, buildMainMenu(h.info.Settings))
}

func (h *handler) sendText(chatID int64, text string) {
	msg := tu.Message(tu.ID(chatID), text)
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
