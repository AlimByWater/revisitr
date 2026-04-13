package masterbot

import (
	"context"
	"fmt"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (h *handler) handleSettings(ctx context.Context, msg *telego.Message) {
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
	if err != nil || len(bots) == 0 {
		h.sendText(msg.Chat.ID, "У вас пока нет ботов. Создайте первого на сайте.")
		return
	}

	// If single bot — show settings directly
	if len(bots) == 1 {
		h.sendBotSettings(ctx, msg.Chat.ID, &bots[0])
		return
	}

	// Multiple bots — ask which one
	var rows [][]telego.InlineKeyboardButton
	for _, b := range bots {
		label := b.Name
		if b.Username != "" {
			label += " (@" + b.Username + ")"
		}
		rows = append(rows, tu.InlineKeyboardRow(
			tu.InlineKeyboardButton(label).WithCallbackData(fmt.Sprintf("settings:%d", b.ID)),
		))
	}

	kb := &telego.InlineKeyboardMarkup{InlineKeyboard: rows}
	tgMsg := tu.Message(tu.ID(msg.Chat.ID), "Выберите бота для настройки:").WithReplyMarkup(kb)
	if _, err := h.bot.SendMessage(context.Background(), tgMsg); err != nil {
		h.logger.Error("send settings bot picker", "error", err)
	}
}

func (h *handler) sendBotSettings(_ context.Context, chatID int64, bot *entity.Bot) {
	welcome := bot.Settings.WelcomeMessage
	if welcome == "" {
		welcome = "(не задано)"
	}
	if len(welcome) > 100 {
		welcome = welcome[:100] + "..."
	}

	formFields := len(bot.Settings.RegistrationForm)
	modules := len(bot.Settings.Modules)

	text := fmt.Sprintf("⚙️ Настройки бота *%s*\n\n"+
		"📝 Welcome: %s\n"+
		"📋 Поля регформы: %d\n"+
		"📦 Модули: %d\n\n"+
		"Для изменения настроек используйте веб-панель.",
		bot.Name, welcome, formFields, modules)

	h.sendText(chatID, text)
}
