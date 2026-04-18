package botmanager

import (
	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	btnLoyalty  = "Лояльность"
	btnMenu     = "Меню"
	btnBooking  = "Забронировать"
	btnFeedback = "Связаться"
	btnContacts = "Контакты"
	btnHome     = "На главную"
	btnAbout    = "ℹ️ О нас"
	btnBack     = "◀️ Назад"
)

func buildMainMenu(settings entity.BotSettings) *telego.ReplyKeyboardMarkup {
	var rows [][]telego.KeyboardButton
	var contentButtons []telego.KeyboardButton

	for _, label := range moduleButtonLabels(settings.Modules) {
		contentButtons = append(contentButtons, tu.KeyboardButton(label))
	}

	if len(settings.Buttons) > 0 {
		for _, btn := range settings.Buttons {
			contentButtons = append(contentButtons, tu.KeyboardButton(btn.Label))
		}
	}

	contentButtons = append(contentButtons, tu.KeyboardButton(btnContacts))

	for len(contentButtons) >= 2 {
		rows = append(rows, []telego.KeyboardButton{contentButtons[0], contentButtons[1]})
		contentButtons = contentButtons[2:]
	}
	if len(contentButtons) == 1 {
		rows = append(rows, []telego.KeyboardButton{contentButtons[0]})
	}
	rows = append(rows, []telego.KeyboardButton{tu.KeyboardButton(btnHome)})

	kb := &telego.ReplyKeyboardMarkup{
		Keyboard:       rows,
		ResizeKeyboard: true,
	}

	return kb
}

func buildContactRequest() *telego.ReplyKeyboardMarkup {
	return &telego.ReplyKeyboardMarkup{
		Keyboard: [][]telego.KeyboardButton{
			{
				{Text: "📱 Отправить номер телефона", RequestContact: true},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}

func buildBackButton() *telego.ReplyKeyboardMarkup {
	return &telego.ReplyKeyboardMarkup{
		Keyboard: [][]telego.KeyboardButton{
			{tu.KeyboardButton(btnBack)},
		},
		ResizeKeyboard: true,
	}
}

func moduleButtonLabels(modules []string) []string {
	seen := map[string]struct{}{}
	var labels []string

	add := func(label string) {
		if _, ok := seen[label]; ok {
			return
		}
		seen[label] = struct{}{}
		labels = append(labels, label)
	}

	for _, module := range modules {
		switch module {
		case "loyalty":
			add(btnLoyalty)
		case "menu":
			add(btnMenu)
		case "booking":
			add(btnBooking)
		case "feedback":
			add(btnFeedback)
		}
	}

	return labels
}
