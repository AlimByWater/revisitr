package botmanager

import (
	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	btnBalance    = "💰 Баланс"
	btnLocations  = "📍 Наши точки"
	btnAbout      = "ℹ️ О нас"
	btnBack       = "◀️ Назад"
)

func buildMainMenu(settings entity.BotSettings) *telego.ReplyKeyboardMarkup {
	var rows [][]telego.KeyboardButton

	// Add custom buttons from settings in pairs
	if len(settings.Buttons) > 0 {
		var row []telego.KeyboardButton
		for _, btn := range settings.Buttons {
			row = append(row, tu.KeyboardButton(btn.Label))
			if len(row) == 2 {
				rows = append(rows, row)
				row = nil
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	// Default buttons
	rows = append(rows, []telego.KeyboardButton{
		tu.KeyboardButton(btnBalance),
		tu.KeyboardButton(btnLocations),
	})

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
