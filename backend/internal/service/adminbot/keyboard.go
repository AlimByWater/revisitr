package adminbot

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	btnStats     = "📊 Статистика"
	btnCampaigns = "📬 Рассылки"
	btnPromos    = "🏷️ Акции"
	btnHelp      = "❓ Помощь"
)

func buildAdminMenu() *telego.ReplyKeyboardMarkup {
	return &telego.ReplyKeyboardMarkup{
		Keyboard: [][]telego.KeyboardButton{
			{tu.KeyboardButton(btnStats), tu.KeyboardButton(btnCampaigns)},
			{tu.KeyboardButton(btnPromos), tu.KeyboardButton(btnHelp)},
		},
		ResizeKeyboard: true,
	}
}
