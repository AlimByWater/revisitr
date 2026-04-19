package botmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"revisitr/internal/entity"
	tgService "revisitr/internal/service/telegram"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type handler struct {
	mgr      *Manager
	bot      *telego.Bot
	info     *entity.Bot // pointer to botInstance.info — updated on hot reload
	tgSender *tgService.Sender
	logger   *slog.Logger
}

func newHandler(mgr *Manager, bot *telego.Bot, info *entity.Bot) *handler {
	return &handler{
		mgr:      mgr,
		bot:      bot,
		info:     info,
		tgSender: mgr.tgSender,
		logger:   mgr.logger.With("bot_id", info.ID, "bot_name", info.Name),
	}
}

func (h *handler) Handle(ctx context.Context, update telego.Update) {
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

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
	state, _ := h.loadFlowState(ctx, chatID)

	if state != nil && state.AwaitingFeedback && text != "" && !strings.HasPrefix(text, "/") {
		h.handleFeedbackResponse(ctx, msg, text, *state)
		return
	}

	switch {
	case text == "/start":
		h.handleStart(ctx, msg)
	case text == btnHome:
		h.handleStart(ctx, msg)
	case text == btnLoyalty:
		h.handleBalance(ctx, msg)
	case text == btnContacts:
		h.handleLocations(ctx, msg)
	case text == btnMenu:
		h.handleMenu(ctx, msg)
	case text == btnBooking:
		h.handleBooking(ctx, msg)
	case text == btnFeedback:
		h.handleFeedback(ctx, msg)
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

	_ = h.clearFlowState(ctx, chatID)

	posIDs, err := h.boundPOSIDs(ctx)
	if err == nil && h.info.Settings.PosSelectorEnabled && len(posIDs) > 1 {
		h.sendVenueChooser(ctx, chatID, posIDs)
		return
	}

	// Check if user already registered
	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, telegramID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		h.logger.Error("check client registration", "error", err, "telegram_id", telegramID)
		h.sendText(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if client != nil {
		// Already registered — show synced welcome content if configured
		if h.hasWelcomeContent() {
			h.sendWelcomeContent(ctx, chatID)
			h.sendMainMenu(ctx, chatID, h.returningMenuText(client.FirstName))
			return
		}

		welcome := h.info.Settings.WelcomeMessage
		if welcome == "" {
			welcome = fmt.Sprintf("С возвращением, %s! 👋", client.FirstName)
		}
		h.sendMainMenu(ctx, chatID, welcome)
		return
	}

	// New user — send welcome content first, then handle registration
	h.sendWelcomeContent(ctx, chatID)

	// If registration form requires phone, ask for it
	needsPhone := false
	for _, field := range h.info.Settings.RegistrationForm {
		if field.Name == "phone" && field.Required {
			needsPhone = true
			break
		}
	}

	if needsPhone {
		h.sendWithKeyboard(chatID, h.registrationPrompt(), buildContactRequest())
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
		h.sendMainMenu(ctx, chatID, h.alreadyRegisteredText())
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

	h.sendMainMenu(ctx, chatID, h.registeredWelcomeText(client.FirstName))
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

	h.sendMainMenu(ctx, chatID, h.registeredWelcomeText(client.FirstName))
}

func (h *handler) handleBalance(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	telegramID := msg.From.ID

	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, telegramID)
	if err != nil {
		h.sendText(chatID, h.balanceNeedsRegistrationText())
		return
	}

	programs, err := h.mgr.loyaltyRepo.GetProgramsByOrgID(ctx, h.info.OrgID)
	if err != nil || len(programs) == 0 {
		h.sendText(chatID, h.loyaltyUnavailableText())
		return
	}

	var sb strings.Builder
	sb.WriteString(h.balanceHeader())

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

		if h.isBaratieDemo() {
			sb.WriteString(fmt.Sprintf("⚓ %s\n", prog.Name))
			sb.WriteString(fmt.Sprintf("   🪙 Дублоны: %.0f %s\n", balance, currencyName))
			sb.WriteString(fmt.Sprintf("   🧭 Ранг экипажа: %s\n\n", levelName))
		} else {
			sb.WriteString(fmt.Sprintf("📋 %s\n", prog.Name))
			sb.WriteString(fmt.Sprintf("   Баланс: %.0f %s\n", balance, currencyName))
			sb.WriteString(fmt.Sprintf("   Уровень: %s\n\n", levelName))
		}
	}

	h.sendText(chatID, sb.String())
}

func (h *handler) handleLocations(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID

	locations, err := h.filteredContactsLocations(ctx, chatID)
	if err != nil || len(locations) == 0 {
		h.sendText(chatID, h.locationsUnavailableText())
		return
	}

	var sb strings.Builder
	sb.WriteString(h.locationsHeader())

	for _, loc := range locations {
		if !loc.IsActive {
			continue
		}
		if h.isBaratieDemo() {
			sb.WriteString(fmt.Sprintf("⚓ %s\n", loc.Name))
			if loc.Address != "" {
				sb.WriteString(fmt.Sprintf("   🪸 %s\n", loc.Address))
			}
			if loc.Phone != "" {
				sb.WriteString(fmt.Sprintf("   📯 %s\n", loc.Phone))
			}
			sb.WriteString("\n")
		} else {
			sb.WriteString(fmt.Sprintf("🏠 %s\n", loc.Name))
			if loc.Address != "" {
				sb.WriteString(fmt.Sprintf("   📌 %s\n", loc.Address))
			}
			if loc.Phone != "" {
				sb.WriteString(fmt.Sprintf("   📞 %s\n", loc.Phone))
			}
			sb.WriteString("\n")
		}
	}

	h.sendText(chatID, sb.String())
}

func (h *handler) handleAbout(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	if h.isBaratieDemo() {
		h.sendText(chatID, "⋆｡°✩ ⚓︎ Baratie ⚓︎ ✩°｡⋆\n\nЛегендарный ресторан на воде, где кухня Sanji встречается с морским ветром East Blue.\n\nОткройте меню, проверьте дублоны и выберите курс к Baratie.")
		return
	}
	h.sendText(chatID, fmt.Sprintf("ℹ️ %s\n\nИспользуйте меню для навигации.", h.info.Name))
}

func (h *handler) handleCustomButton(ctx context.Context, msg *telego.Message, text string) {
	for idx, btn := range h.info.Settings.Buttons {
		if btn.Label == text {
			if btn.Content != nil && len(btn.Content.Parts) > 0 {
				if h.tgSender != nil {
					if updated, changed, err := h.tgSender.SendContentWithCache(ctx, h.bot, msg.Chat.ID, *btn.Content); err == nil {
						if changed {
							h.info.Settings.Buttons = h.updateButtonContentCache(idx, updated)
							h.persistSettingsCache(ctx)
						}
						return
					} else {
						h.logger.Error("send button content", "error", err, "label", btn.Label)
					}
				}

				for _, part := range btn.Content.Parts {
					if part.Text != "" {
						h.sendText(msg.Chat.ID, part.Text)
						return
					}
				}
			}

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

		h.sendText(client.TelegramID, h.welcomeBonusText(bonus, currencyName))
		h.logger.Info("welcome bonus awarded", "client_id", client.ID, "amount", bonus, "program", prog.Name)
	}
}

func (h *handler) sendMainMenu(_ context.Context, chatID int64, text string) {
	h.sendWithKeyboard(chatID, text, buildMainMenu(h.info.Settings))
}

// sendWelcomeContent sends the composite welcome message if available.
func (h *handler) sendWelcomeContent(ctx context.Context, chatID int64) {
	settings := h.info.Settings

	// Priority: new format → legacy → default
	if h.tgSender != nil && h.hasWelcomeContent() {
		if updated, changed, err := h.tgSender.SendContentWithCache(ctx, h.bot, chatID, *settings.WelcomeContent); err != nil {
			h.logger.Error("send welcome content", "error", err, "chat_id", chatID)
		} else if changed {
			h.info.Settings.WelcomeContent = &updated
			h.persistSettingsCache(ctx)
		}
		return
	}

	// Legacy text welcome
	welcome := settings.WelcomeMessage
	if welcome == "" {
		if h.isBaratieDemo() {
			welcome = "⋆｡°✩ ⚓︎ Добро пожаловать на борт Baratie ⚓︎ ✩°｡⋆"
		} else {
			welcome = fmt.Sprintf("Добро пожаловать в %s! 🎉", h.info.Name)
		}
	}
	h.sendText(chatID, welcome)
}

func (h *handler) hasWelcomeContent() bool {
	return h.info.Settings.WelcomeContent != nil && len(h.info.Settings.WelcomeContent.Parts) > 0
}

func (h *handler) isBaratieDemo() bool {
	return h.info.Username == "baratie_demo_bot"
}

func (h *handler) registrationPrompt() string {
	if h.isBaratieDemo() {
		return "╭──────༺🪸༻──────╮\nПоделитесь номером телефона,\nи мы впишем вас в список гостей Baratie.\n╰──────༺🦈༻──────╯"
	}
	return "Для регистрации поделитесь номером телефона:"
}

func (h *handler) alreadyRegisteredText() string {
	if h.isBaratieDemo() {
		return "⚓ Вы уже в списке гостей Baratie. Добро пожаловать на борт!"
	}
	return "Вы уже зарегистрированы! 👍"
}

func (h *handler) registeredWelcomeText(firstName string) string {
	if h.isBaratieDemo() {
		return fmt.Sprintf("⚓ %s, вы успешно поднялись на борт Baratie!\n\nОткройте меню, проверьте дублоны и выберите свой курс.", firstName)
	}
	return fmt.Sprintf("Добро пожаловать, %s! Вы успешно зарегистрированы. 🎉", firstName)
}

func (h *handler) returningMenuText(firstName string) string {
	if h.isBaratieDemo() {
		return fmt.Sprintf("⚓ %s, экипаж Baratie к вашим услугам. Выберите курс в меню ниже.", firstName)
	}
	return "Главное меню"
}

func (h *handler) balanceNeedsRegistrationText() string {
	if h.isBaratieDemo() {
		return "⚓ Сначала поднимитесь на борт через /start, чтобы открыть трюм с дублонами."
	}
	return "Сначала нужно зарегистрироваться. Отправьте /start"
}

func (h *handler) loyaltyUnavailableText() string {
	if h.isBaratieDemo() {
		return "🪙 Корабельная казна пока недоступна. Попробуйте чуть позже."
	}
	return "Программа лояльности пока не настроена."
}

func (h *handler) balanceHeader() string {
	if h.isBaratieDemo() {
		return "⋆｡°✩ 🪙 Трюм дублонов Baratie 🪙 ✩°｡⋆\n\n"
	}
	return "💰 Ваш баланс:\n\n"
}

func (h *handler) locationsUnavailableText() string {
	if h.isBaratieDemo() {
		return "🧭 Курс к Baratie пока уточняется. Загляните чуть позже."
	}
	return "Информация о точках пока недоступна."
}

func (h *handler) locationsHeader() string {
	if h.isBaratieDemo() {
		return "⋆｡°✩ 🧭 Курс к Baratie 🧭 ✩°｡⋆\n\n"
	}
	return "📍 Контакты:\n\n"
}

func (h *handler) welcomeBonusText(bonus float64, currencyName string) string {
	if h.isBaratieDemo() {
		return fmt.Sprintf("🪙 В ваш трюм зачислено %.0f %s!\nДобро пожаловать в экипаж Baratie.", bonus, currencyName)
	}
	return fmt.Sprintf("🎁 Вам начислен приветственный бонус: %.0f %s!", bonus, currencyName)
}

func (h *handler) sendText(chatID int64, text string) {
	msg := tu.Message(tu.ID(chatID), stripEmojiMarkers(text))
	if _, err := h.bot.SendMessage(context.Background(), msg); err != nil {
		h.logger.Error("send message", "error", err, "chat_id", chatID)
	}
}

func (h *handler) sendWithKeyboard(chatID int64, text string, kb *telego.ReplyKeyboardMarkup) {
	msg := tu.Message(tu.ID(chatID), stripEmojiMarkers(text)).WithReplyMarkup(kb)
	if _, err := h.bot.SendMessage(context.Background(), msg); err != nil {
		h.logger.Error("send message with keyboard", "error", err, "chat_id", chatID)
	}
}

func (h *handler) updateButtonContentCache(index int, updated entity.MessageContent) []entity.BotButton {
	buttons := append([]entity.BotButton(nil), h.info.Settings.Buttons...)
	if index >= 0 && index < len(buttons) {
		buttons[index].Content = &updated
	}
	return buttons
}

func (h *handler) persistSettingsCache(ctx context.Context) {
	if err := h.mgr.botsRepo.UpdateSettings(ctx, h.info.ID, h.info.Settings); err != nil {
		h.logger.Error("persist cached media ids", "error", err, "bot_id", h.info.ID)
	}
}
