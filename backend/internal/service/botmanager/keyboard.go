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
	var currentRow []telego.KeyboardButton

	for _, button := range orderedMainMenuButtons(settings) {
		if button.Label == btnHome {
			if len(currentRow) > 0 {
				rows = append(rows, currentRow)
				currentRow = nil
			}
			rows = append(rows, []telego.KeyboardButton{tu.KeyboardButton(button.Label)})
			continue
		}

		currentRow = append(currentRow, tu.KeyboardButton(button.Label))
		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) == 1 {
		rows = append(rows, currentRow)
	}

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

func orderedMainMenuButtons(settings entity.BotSettings) []entity.BotButton {
	required := map[string]entity.BotButton{}

	addRequired := func(label string, managedByModule *string, isSystem bool) {
		required[label] = entity.BotButton{
			Label:           label,
			Type:            "text",
			Value:           label,
			ManagedByModule: managedByModule,
			IsSystem:        isSystem,
		}
	}

	for _, module := range settings.Modules {
		switch module {
		case "loyalty":
			moduleName := "loyalty"
			addRequired(btnLoyalty, &moduleName, true)
		case "menu":
			moduleName := "menu"
			addRequired(btnMenu, &moduleName, true)
		case "booking":
			moduleName := "booking"
			addRequired(btnBooking, &moduleName, true)
		case "feedback":
			moduleName := "feedback"
			addRequired(btnFeedback, &moduleName, true)
		}
	}

	contacts := "contacts"
	home := "home"
	addRequired(btnContacts, &contacts, true)
	addRequired(btnHome, &home, true)

	seen := map[string]struct{}{}
	var ordered []entity.BotButton

	for _, button := range settings.Buttons {
		if _, ok := required[button.Label]; button.IsSystem || button.ManagedByModule != nil || ok {
			if _, exists := required[button.Label]; !exists {
				continue
			}
			ordered = append(ordered, required[button.Label])
			seen[button.Label] = struct{}{}
			continue
		}

		ordered = append(ordered, button)
		seen[button.Label] = struct{}{}
	}

	for _, label := range moduleButtonLabels(settings.Modules) {
		if _, ok := seen[label]; ok {
			continue
		}
		ordered = append(ordered, required[label])
		seen[label] = struct{}{}
	}
	for _, label := range []string{btnContacts, btnHome} {
		if _, ok := seen[label]; ok {
			continue
		}
		ordered = append(ordered, required[label])
		seen[label] = struct{}{}
	}

	return ordered
}
