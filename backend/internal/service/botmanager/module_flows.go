package botmanager

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"time"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	callbackVenuePrefix      = "venue:"
	callbackMenuRoot         = "menu:root"
	callbackMenuCategoryPref = "menu:cat:"
	callbackMenuItemPref     = "menu:item:"
	callbackBookingIntro     = "booking:intro"
	callbackBookingStart     = "booking:start"
	callbackBookingDatePage  = "booking:date-page:"
	callbackBookingPickDate  = "booking:pick-date:"
	callbackBookingTimePage  = "booking:time-page:"
	callbackBookingPickTime  = "booking:pick-time:"
	callbackBookingPartyPage = "booking:party-page"
	callbackBookingPickParty = "booking:pick-party:"
)

func (h *handler) loadFlowState(ctx context.Context, chatID int64) (*FlowState, error) {
	if h.mgr.sessions == nil {
		return &FlowState{}, nil
	}
	return h.mgr.sessions.Load(ctx, h.info.ID, chatID)
}

func (h *handler) saveFlowState(ctx context.Context, chatID int64, state FlowState) error {
	if h.mgr.sessions == nil {
		return nil
	}
	return h.mgr.sessions.Save(ctx, h.info.ID, chatID, state)
}

func (h *handler) clearFlowState(ctx context.Context, chatID int64) error {
	if h.mgr.sessions == nil {
		return nil
	}
	return h.mgr.sessions.Delete(ctx, h.info.ID, chatID)
}

func (h *handler) currentFlowState(ctx context.Context, chatID int64) FlowState {
	state, err := h.loadFlowState(ctx, chatID)
	if err != nil {
		h.logger.Error("load flow state", "error", err, "chat_id", chatID)
		return FlowState{}
	}
	if state == nil {
		return FlowState{}
	}
	return *state
}

func (h *handler) promptPOSSelection(ctx context.Context, chatID int64) bool {
	if _, ok := h.resolveSelectedPOS(ctx, chatID, false); ok || !h.info.Settings.PosSelectorEnabled {
		return false
	}

	posIDs, err := h.boundPOSIDs(ctx)
	if err != nil || len(posIDs) <= 1 {
		return false
	}

	h.sendVenueChooser(ctx, chatID, posIDs)
	return true
}

func (h *handler) menuUnavailableMessage() string {
	if message := h.info.Settings.ModuleConfigs.Menu.UnavailableMessage; message != "" {
		return message
	}
	return "Меню пока недоступно."
}

func (h *handler) feedbackPromptMessage() string {
	if prompt := h.info.Settings.ModuleConfigs.Feedback.PromptMessage; prompt != "" {
		return prompt
	}
	return "Напишите ваш вопрос:"
}

func (h *handler) feedbackSuccessMessage() string {
	if success := h.info.Settings.ModuleConfigs.Feedback.SuccessMessage; success != "" {
		return success
	}
	return "Ваше сообщение отправлено."
}

func (h *handler) boundPOSIDs(ctx context.Context) ([]int, error) {
	if h.mgr.menusRepo == nil {
		return nil, nil
	}
	return h.mgr.menusRepo.GetBotPOSLocations(ctx, h.info.ID)
}

func (h *handler) resolveSelectedPOS(ctx context.Context, chatID int64, forceChooser bool) (int, bool) {
	posIDs, err := h.boundPOSIDs(ctx)
	if err != nil {
		h.logger.Error("resolve bound pos ids", "error", err)
		return 0, false
	}
	if len(posIDs) == 0 {
		return 0, true
	}
	state, err := h.loadFlowState(ctx, chatID)
	if err != nil {
		h.logger.Error("load flow state", "error", err)
		return 0, false
	}
	if !forceChooser && state != nil && state.SelectedPOSID != 0 {
		for _, posID := range posIDs {
			if posID == state.SelectedPOSID {
				return posID, true
			}
		}
	}
	if len(posIDs) == 1 {
		if state != nil {
			state.SelectedPOSID = posIDs[0]
			_ = h.saveFlowState(ctx, chatID, *state)
		}
		return posIDs[0], true
	}
	return 0, false
}

func (h *handler) sendVenueChooser(ctx context.Context, chatID int64, posIDs []int) {
	locations, err := h.mgr.posRepo.GetByOrgID(ctx, h.info.OrgID)
	if err != nil {
		h.sendText(chatID, "Не удалось загрузить список заведений.")
		return
	}

	locationMap := make(map[int]entity.POSLocation, len(locations))
	for _, location := range locations {
		locationMap[location.ID] = location
	}

	var buttons [][]entity.InlineButton
	for _, posID := range posIDs {
		location, ok := locationMap[posID]
		if !ok || !location.IsActive {
			continue
		}
		buttons = append(buttons, []entity.InlineButton{{
			Text: location.Name,
			Data: callbackVenuePrefix + strconv.Itoa(posID),
		}})
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "venue"
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "Выберите заведение:",
		}},
		Buttons: buttons,
	})
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleCallbackQuery(ctx context.Context, query *telego.CallbackQuery) {
	if query == nil || query.Message == nil {
		return
	}
	chatID := query.Message.GetChat().ID
	data := query.Data

	switch {
	case strings.HasPrefix(data, callbackVenuePrefix):
		h.answerCallback(query.ID, "")
		h.handleVenueSelection(ctx, query, strings.TrimPrefix(data, callbackVenuePrefix))
	case data == callbackMenuRoot:
		h.answerCallback(query.ID, "")
		h.renderMenuRoot(ctx, chatID)
	case strings.HasPrefix(data, callbackMenuCategoryPref):
		h.answerCallback(query.ID, "")
		h.handleMenuCategoryCallback(ctx, chatID, strings.TrimPrefix(data, callbackMenuCategoryPref))
	case strings.HasPrefix(data, callbackMenuItemPref):
		h.answerCallback(query.ID, "")
		h.handleMenuItemCallback(ctx, chatID, strings.TrimPrefix(data, callbackMenuItemPref))
	case data == callbackBookingIntro:
		h.answerCallback(query.ID, "")
		h.renderBookingIntro(ctx, chatID)
	case data == callbackBookingStart:
		h.answerCallback(query.ID, "")
		h.renderBookingDates(ctx, chatID, 0)
	case strings.HasPrefix(data, callbackBookingDatePage):
		h.answerCallback(query.ID, "")
		page, _ := strconv.Atoi(strings.TrimPrefix(data, callbackBookingDatePage))
		h.renderBookingDates(ctx, chatID, page)
	case strings.HasPrefix(data, callbackBookingPickDate):
		h.answerCallback(query.ID, "")
		h.handleBookingDateSelection(ctx, chatID, strings.TrimPrefix(data, callbackBookingPickDate))
	case strings.HasPrefix(data, callbackBookingTimePage):
		h.answerCallback(query.ID, "")
		page, _ := strconv.Atoi(strings.TrimPrefix(data, callbackBookingTimePage))
		h.renderBookingTimes(ctx, chatID, page)
	case strings.HasPrefix(data, callbackBookingPickTime):
		h.answerCallback(query.ID, "")
		h.handleBookingTimeSelection(ctx, chatID, strings.TrimPrefix(data, callbackBookingPickTime))
	case data == callbackBookingPartyPage:
		h.answerCallback(query.ID, "")
		h.renderBookingPartySizes(ctx, chatID)
	case strings.HasPrefix(data, callbackBookingPickParty):
		h.answerCallback(query.ID, "")
		h.handleBookingPartySelection(ctx, chatID, strings.TrimPrefix(data, callbackBookingPickParty))
	default:
		h.answerCallback(query.ID, "Действие устарело")
	}
}

func (h *handler) handleVenueSelection(ctx context.Context, query *telego.CallbackQuery, value string) {
	posID, err := strconv.Atoi(value)
	if err != nil {
		return
	}
	chatID := query.Message.GetChat().ID
	state := h.currentFlowState(ctx, chatID)
	state.SelectedPOSID = posID
	state.Flow = ""
	state.FlowMessageID = 0
	_ = h.saveFlowState(ctx, chatID, state)
	_ = h.deleteMessage(chatID, query.Message.GetMessageID())
	h.sendWelcomeAndMenuAfterStart(ctx, chatID, &query.From)
}

func (h *handler) sendWelcomeAndMenuAfterStart(ctx context.Context, chatID int64, user *telego.User) {
	if user == nil {
		return
	}
	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, user.ID)
	if err != nil && !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		h.logger.Error("check client registration", "error", err, "telegram_id", user.ID)
		h.sendText(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if client != nil {
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

	h.sendWelcomeContent(ctx, chatID)

	needsPhone := false
	for _, field := range h.info.Settings.RegistrationForm {
		if field.Name == "phone" && field.Required {
			needsPhone = true
			break
		}
	}

	if needsPhone {
		h.sendWithKeyboard(chatID, h.registrationPrompt(), buildContactRequest())
		return
	}

	msg := &telego.Message{
		Chat:      telego.Chat{ID: chatID},
		From:      user,
		MessageID: 0,
	}
	h.autoRegister(ctx, msg)
}

func (h *handler) filteredContactsLocations(ctx context.Context, chatID int64) ([]entity.POSLocation, error) {
	locations, err := h.mgr.posRepo.GetByOrgID(ctx, h.info.OrgID)
	if err != nil {
		return nil, err
	}

	boundPOSIDs, err := h.boundPOSIDs(ctx)
	if err != nil {
		return nil, err
	}

	allowed := make(map[int]struct{})
	sourceIDs := h.info.Settings.ContactsPOSIDs
	if len(sourceIDs) == 0 {
		sourceIDs = boundPOSIDs
	}
	if len(sourceIDs) == 0 {
		return nil, nil
	}
	for _, posID := range sourceIDs {
		allowed[posID] = struct{}{}
	}

	var filtered []entity.POSLocation
	for _, location := range locations {
		if !location.IsActive {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[location.ID]; !ok {
				continue
			}
		}
		filtered = append(filtered, location)
	}

	state := h.currentFlowState(ctx, chatID)
	selected := 0
	selected = state.SelectedPOSID
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].ID == selected {
			return true
		}
		if filtered[j].ID == selected {
			return false
		}
		return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
	})

	return filtered, nil
}

func (h *handler) handleMenu(ctx context.Context, msg *telego.Message) {
	if h.promptPOSSelection(ctx, msg.Chat.ID) {
		return
	}
	h.renderMenuRoot(ctx, msg.Chat.ID)
}

func (h *handler) renderMenuRoot(ctx context.Context, chatID int64) {
	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	var buttons [][]entity.InlineButton
	for _, category := range menu.Categories {
		label := category.Name
		if category.IconEmoji != nil && *category.IconEmoji != "" {
			label = *category.IconEmoji + " " + label
		}
		buttons = append(buttons, []entity.InlineButton{{
			Text: label,
			Data: callbackMenuCategoryPref + strconv.Itoa(category.ID) + ":0",
		}})
	}

	content := h.menuIntroContent(menu)
	content.Buttons = buttons

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu"
	state.MenuCategoryID = 0
	state.MenuPage = 0
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleMenuCategoryCallback(ctx context.Context, chatID int64, value string) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return
	}
	categoryID, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	page, _ := strconv.Atoi(parts[1])

	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	var category *entity.MenuCategory
	for i := range menu.Categories {
		if menu.Categories[i].ID == categoryID {
			category = &menu.Categories[i]
			break
		}
	}
	if category == nil {
		return
	}

	pageSize := 9
	start := page * pageSize
	if start > len(category.Items) {
		start = 0
		page = 0
	}
	end := start + pageSize
	if end > len(category.Items) {
		end = len(category.Items)
	}

	var buttons [][]entity.InlineButton
	for _, item := range category.Items[start:end] {
		buttons = append(buttons, []entity.InlineButton{{
			Text: item.Name,
			Data: callbackMenuItemPref + strconv.Itoa(item.ID) + ":" + strconv.Itoa(category.ID) + ":" + strconv.Itoa(page),
		}})
	}

	if len(category.Items) > pageSize {
		var nav []entity.InlineButton
		if page > 0 {
			nav = append(nav, entity.InlineButton{Text: "←", Data: callbackMenuCategoryPref + strconv.Itoa(category.ID) + ":" + strconv.Itoa(page-1)})
		}
		if end < len(category.Items) {
			nav = append(nav, entity.InlineButton{Text: "→", Data: callbackMenuCategoryPref + strconv.Itoa(category.ID) + ":" + strconv.Itoa(page+1)})
		}
		if len(nav) > 0 {
			buttons = append(buttons, nav)
		}
	}
	buttons = append(buttons, []entity.InlineButton{{Text: "Назад", Data: callbackMenuRoot}})

	content := entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: category.Name + "\n\nВыберите позицию:",
		}},
		Buttons: buttons,
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu"
	state.MenuCategoryID = categoryID
	state.MenuPage = page
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleMenuItemCallback(ctx context.Context, chatID int64, value string) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return
	}
	itemID, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	categoryID, _ := strconv.Atoi(parts[1])
	page, _ := strconv.Atoi(parts[2])

	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	for _, category := range menu.Categories {
		if category.ID != categoryID {
			continue
		}
		for _, item := range category.Items {
			if item.ID != itemID {
				continue
			}

			text := "<b>" + html.EscapeString(item.Name) + "</b>"
			if item.Description != nil && *item.Description != "" {
				text += "\n\n" + html.EscapeString(*item.Description)
			}
			text += fmt.Sprintf("\n\nЦена: %.0f ₽", item.Price)
			if item.Weight != nil && *item.Weight != "" {
				text += "\nГраммаж: " + html.EscapeString(*item.Weight)
			}

			content := entity.MessageContent{
				Parts: []entity.MessagePart{{
					Type:      entity.PartText,
					Text:      text,
					ParseMode: "HTML",
				}},
				Buttons: [][]entity.InlineButton{{
					{Text: "Назад", Data: callbackMenuCategoryPref + strconv.Itoa(categoryID) + ":" + strconv.Itoa(page)},
				}},
			}
			if item.ImageURL != nil && *item.ImageURL != "" {
				content.Parts[0] = entity.MessagePart{
					Type:      entity.PartPhoto,
					MediaURL:  *item.ImageURL,
					Text:      text,
					ParseMode: "HTML",
				}
			}

			state := h.currentFlowState(ctx, chatID)
			state.Flow = "menu"
			state.MenuCategoryID = categoryID
			state.MenuPage = page
			state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
			_ = h.saveFlowState(ctx, chatID, state)
			return
		}
	}
}

func (h *handler) handleBooking(ctx context.Context, msg *telego.Message) {
	if h.promptPOSSelection(ctx, msg.Chat.ID) {
		return
	}
	if !h.bookingAvailableForChat(ctx, msg.Chat.ID) {
		h.sendText(msg.Chat.ID, "Бронирование для выбранной точки продаж сейчас недоступно.")
		return
	}
	h.renderBookingIntro(ctx, msg.Chat.ID)
}

func (h *handler) renderBookingIntro(ctx context.Context, chatID int64) {
	if !h.bookingAvailableForChat(ctx, chatID) {
		h.sendText(chatID, "Бронирование для выбранной точки продаж сейчас недоступно.")
		return
	}
	state := h.currentFlowState(ctx, chatID)
	state.Flow = "booking"
	state.BookingStage = "intro"

	content := entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "Забронировать столик можно, нажав кнопку ниже.",
		}},
		Buttons: [][]entity.InlineButton{{
			{Text: "Забронировать в боте", Data: callbackBookingStart},
		}},
	}
	if intro := h.info.Settings.ModuleConfigs.Booking.IntroContent; intro != nil && len(intro.Parts) > 0 {
		content = *intro
		content.Buttons = [][]entity.InlineButton{{
			{Text: "Забронировать в боте", Data: callbackBookingStart},
		}}
	}

	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) renderBookingDates(ctx context.Context, chatID int64, page int) {
	state := h.currentFlowState(ctx, chatID)

	dates := h.availableBookingDates()
	pageSize := 16
	start := page * pageSize
	if start > len(dates) {
		start = 0
		page = 0
	}
	end := start + pageSize
	if end > len(dates) {
		end = len(dates)
	}

	var buttons [][]entity.InlineButton
	for i := start; i < end; i += 2 {
		row := []entity.InlineButton{{
			Text: dates[i].Format("02.01"),
			Data: callbackBookingPickDate + dates[i].Format("2006-01-02") + ":" + strconv.Itoa(page),
		}}
		if i+1 < end {
			row = append(row, entity.InlineButton{
				Text: dates[i+1].Format("02.01"),
				Data: callbackBookingPickDate + dates[i+1].Format("2006-01-02") + ":" + strconv.Itoa(page),
			})
		}
		buttons = append(buttons, row)
	}

	var nav []entity.InlineButton
	if page > 0 {
		nav = append(nav, entity.InlineButton{Text: "←", Data: callbackBookingDatePage + strconv.Itoa(page-1)})
	}
	if end < len(dates) {
		nav = append(nav, entity.InlineButton{Text: "→", Data: callbackBookingDatePage + strconv.Itoa(page+1)})
	}
	if len(nav) > 0 {
		buttons = append(buttons, nav)
	}
	buttons = append(buttons, []entity.InlineButton{{Text: "Назад", Data: callbackBookingIntro}})

	content := entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "Выберите дату:",
		}},
		Buttons: buttons,
	}

	state.Flow = "booking"
	state.BookingStage = "date"
	state.BookingPage = page
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleBookingDateSelection(ctx context.Context, chatID int64, value string) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return
	}
	state := h.currentFlowState(ctx, chatID)
	state.BookingDate = parts[0]
	page, _ := strconv.Atoi(parts[1])
	state.BookingPage = page
	_ = h.saveFlowState(ctx, chatID, state)
	h.renderBookingTimes(ctx, chatID, 0)
}

func (h *handler) renderBookingTimes(ctx context.Context, chatID int64, page int) {
	state := h.currentFlowState(ctx, chatID)

	slots := h.availableBookingSlots(ctx, chatID)
	pageSize := 8
	start := page * pageSize
	if start > len(slots) {
		start = 0
		page = 0
	}
	end := start + pageSize
	if end > len(slots) {
		end = len(slots)
	}

	var buttons [][]entity.InlineButton
	for index := start; index < end; index++ {
		buttons = append(buttons, []entity.InlineButton{{
			Text: slots[index],
			Data: callbackBookingPickTime + strconv.Itoa(index),
		}})
	}
	var nav []entity.InlineButton
	if page > 0 {
		nav = append(nav, entity.InlineButton{Text: "←", Data: callbackBookingTimePage + strconv.Itoa(page-1)})
	}
	if end < len(slots) {
		nav = append(nav, entity.InlineButton{Text: "→", Data: callbackBookingTimePage + strconv.Itoa(page+1)})
	}
	if len(nav) > 0 {
		buttons = append(buttons, nav)
	}
	buttons = append(buttons, []entity.InlineButton{{Text: "Назад", Data: callbackBookingDatePage + strconv.Itoa(state.BookingPage)}})

	content := entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "Выберите время:",
		}},
		Buttons: buttons,
	}

	state.Flow = "booking"
	state.BookingStage = "time"
	state.BookingPage = page
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleBookingTimeSelection(ctx context.Context, chatID int64, value string) {
	index, err := strconv.Atoi(value)
	if err != nil {
		return
	}
	slots := h.availableBookingSlots(ctx, chatID)
	if index < 0 || index >= len(slots) {
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.BookingTime = slots[index]
	_ = h.saveFlowState(ctx, chatID, state)
	h.renderBookingPartySizes(ctx, chatID)
}

func (h *handler) renderBookingPartySizes(ctx context.Context, chatID int64) {
	state := h.currentFlowState(ctx, chatID)
	options := h.bookingPartyOptions()
	var buttons [][]entity.InlineButton
	for index, option := range options {
		buttons = append(buttons, []entity.InlineButton{{
			Text: option,
			Data: callbackBookingPickParty + strconv.Itoa(index),
		}})
	}
	buttons = append(buttons, []entity.InlineButton{{Text: "Назад", Data: callbackBookingTimePage + strconv.Itoa(state.BookingPage)}})

	content := entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "Количество посетителей:",
		}},
		Buttons: buttons,
	}

	state.Flow = "booking"
	state.BookingStage = "party"
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleBookingPartySelection(ctx context.Context, chatID int64, value string) {
	index, err := strconv.Atoi(value)
	if err != nil {
		return
	}
	options := h.bookingPartyOptions()
	if index < 0 || index >= len(options) {
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.BookingPartySize = options[index]
	state.Flow = "booking"
	state.BookingStage = "confirm"

	content := entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: fmt.Sprintf("Бронирование готово:\n\nДата: %s\nВремя: %s\nГости: %s", state.BookingDate, state.BookingTime, state.BookingPartySize),
		}},
		Buttons: [][]entity.InlineButton{{
			{Text: "Назад", Data: callbackBookingPartyPage},
		}},
	}

	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleFeedback(ctx context.Context, msg *telego.Message) {
	state := h.currentFlowState(ctx, msg.Chat.ID)
	state.Flow = "feedback"
	state.AwaitingFeedback = true

	state.FlowMessageID = h.replaceFlowMessage(ctx, msg.Chat.ID, state, entity.MessageContent{
		Parts: []entity.MessagePart{{Type: entity.PartText, Text: h.feedbackPromptMessage()}},
	})
	_ = h.saveFlowState(ctx, msg.Chat.ID, state)
}

func (h *handler) handleFeedbackResponse(ctx context.Context, msg *telego.Message, text string, state FlowState) {
	delivered := false
	if h.mgr.adminBot != nil && h.info.CreatedByTelegramID != nil {
		posName := ""
		if location, ok := h.selectedPOSLocation(ctx, msg.Chat.ID); ok {
			posName = location.Name
		}
		adminText := fmt.Sprintf(
			"Новое сообщение от пользователя <a href=\"tg://user?id=%d\">%s</a>\n\nБот: @%s\n",
			msg.From.ID,
			html.EscapeString(h.userDisplayName(msg.From)),
			html.EscapeString(h.info.Username),
		)
		if posName != "" {
			adminText += "Точка: " + html.EscapeString(posName) + "\n"
		}
		adminText += "\nСообщение:\n" + html.EscapeString(text)
		adminMsg := tu.Message(tu.ID(*h.info.CreatedByTelegramID), adminText).WithParseMode("HTML")
		if _, err := h.mgr.adminBot.SendMessage(ctx, adminMsg); err != nil {
			h.logger.Error("forward feedback to admin bot", "error", err, "bot_id", h.info.ID)
		} else {
			delivered = true
		}
	}

	state.AwaitingFeedback = false
	state.Flow = ""
	state.FlowMessageID = 0
	_ = h.saveFlowState(ctx, msg.Chat.ID, state)
	if delivered {
		h.sendText(msg.Chat.ID, h.feedbackSuccessMessage())
		return
	}
	h.sendText(msg.Chat.ID, "Не удалось отправить сообщение. Попробуйте позже.")
}

func (h *handler) resolveMenuForChat(ctx context.Context, chatID int64) (*entity.Menu, int, bool) {
	if h.mgr.menusRepo == nil {
		return nil, 0, false
	}
	posID, ok := h.resolveSelectedPOS(ctx, chatID, false)
	if !ok || posID == 0 {
		return nil, 0, false
	}
	menu, err := h.mgr.menusRepo.GetActiveMenuForPOS(ctx, h.info.OrgID, posID)
	if err != nil {
		h.logger.Error("get active menu for pos", "error", err, "pos_id", posID, "bot_id", h.info.ID)
		return nil, 0, false
	}
	return menu, posID, menu != nil
}

func (h *handler) menuIntroContent(menu *entity.Menu) entity.MessageContent {
	// Layer 1: Per-menu IntroContent (per-POS override)
	if menu != nil && menu.IntroContent != nil && len(menu.IntroContent.Parts) > 0 {
		return *menu.IntroContent
	}
	// Layer 2: BotButton.Content from General tab
	for _, btn := range h.info.Settings.Buttons {
		if btn.ManagedByModule != nil && *btn.ManagedByModule == "menu" &&
			btn.Content != nil && len(btn.Content.Parts) > 0 {
			return *btn.Content
		}
	}
	// Layer 3: Hardcoded fallback
	text := "Выберите категорию:"
	if menu != nil && menu.Name != "" {
		text = menu.Name + "\n\nВыберите категорию:"
	}
	return entity.MessageContent{
		Parts: []entity.MessagePart{{Type: entity.PartText, Text: text}},
	}
}

func (h *handler) availableBookingDates() []time.Time {
	from := h.info.Settings.ModuleConfigs.Booking.DateFromDays
	to := h.info.Settings.ModuleConfigs.Booking.DateToDays
	if to < from {
		to = from
	}
	if to > 32 {
		to = 32
	}

	today := time.Now()
	var dates []time.Time
	for offset := from; offset <= to; offset++ {
		dates = append(dates, today.AddDate(0, 0, offset))
	}
	return dates
}

func (h *handler) availableBookingSlots(ctx context.Context, chatID int64) []string {
	if !h.bookingAvailableForChat(ctx, chatID) {
		return nil
	}
	configSlots := h.info.Settings.ModuleConfigs.Booking.TimeSlots
	if len(configSlots) > 0 {
		slots := make([]string, 0, len(configSlots))
		for _, slot := range configSlots {
			slots = append(slots, slot.Start+"-"+slot.End)
		}
		return slots
	}

	location, ok := h.selectedPOSLocation(ctx, chatID)
	if !ok {
		return defaultHourlySlots("10:00", "20:00")
	}

	for _, day := range location.Schedule {
		if day.Closed || day.Open == "" || day.Close == "" {
			continue
		}
		return defaultHourlySlots(day.Open, day.Close)
	}

	return defaultHourlySlots("10:00", "20:00")
}

func (h *handler) bookingPartyOptions() []string {
	options := h.info.Settings.ModuleConfigs.Booking.PartySizeOptions
	if len(options) == 0 {
		return []string{"1", "2", "3-5", "6+"}
	}
	return options
}

func (h *handler) bookingAvailableForChat(ctx context.Context, chatID int64) bool {
	configuredPOSIDs := h.info.Settings.ModuleConfigs.Booking.POSIDs
	if len(configuredPOSIDs) == 0 {
		return true
	}
	posID, ok := h.resolveSelectedPOS(ctx, chatID, false)
	if !ok {
		return false
	}
	for _, configuredID := range configuredPOSIDs {
		if configuredID == posID {
			return true
		}
	}
	return false
}

func (h *handler) selectedPOSLocation(ctx context.Context, chatID int64) (entity.POSLocation, bool) {
	posID, ok := h.resolveSelectedPOS(ctx, chatID, false)
	if !ok || posID == 0 {
		return entity.POSLocation{}, false
	}
	locations, err := h.mgr.posRepo.GetByOrgID(ctx, h.info.OrgID)
	if err != nil {
		return entity.POSLocation{}, false
	}
	for _, location := range locations {
		if location.ID == posID {
			return location, true
		}
	}
	return entity.POSLocation{}, false
}

func (h *handler) replaceFlowMessage(ctx context.Context, chatID int64, state FlowState, content entity.MessageContent) int {
	if state.FlowMessageID != 0 {
		_ = h.deleteMessage(chatID, state.FlowMessageID)
	}
	messageID := h.sendContentMessage(ctx, chatID, content)
	return messageID
}

func (h *handler) sendContentMessage(ctx context.Context, chatID int64, content entity.MessageContent) int {
	if len(content.Parts) == 0 {
		return 0
	}

	part := content.Parts[0]
	markup := h.inlineKeyboard(content.Buttons)

	switch part.Type {
	case entity.PartPhoto:
		photo := tu.Photo(tu.ID(chatID), h.inputFile(part))
		if part.Text != "" {
			photo = photo.WithCaption(part.Text)
		}
		if part.ParseMode != "" {
			photo = photo.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			photo = photo.WithReplyMarkup(markup)
		}
		message, err := h.bot.SendPhoto(ctx, photo)
		if err != nil {
			h.logger.Error("send flow photo", "error", err, "chat_id", chatID)
			return 0
		}
		return message.MessageID
	default:
		msg := tu.Message(tu.ID(chatID), part.Text)
		if part.ParseMode != "" {
			msg = msg.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			msg = msg.WithReplyMarkup(markup)
		}
		message, err := h.bot.SendMessage(ctx, msg)
		if err != nil {
			h.logger.Error("send flow text", "error", err, "chat_id", chatID)
			return 0
		}
		return message.MessageID
	}
}

func (h *handler) inputFile(part entity.MessagePart) telego.InputFile {
	if part.MediaID != "" {
		return tu.FileFromID(part.MediaID)
	}
	mediaURL := part.MediaURL
	if mediaURL == "" {
		return tu.FileFromURL("https://elysium.fm")
	}
	if strings.HasPrefix(mediaURL, "http") {
		return tu.FileFromURL(mediaURL)
	}
	return tu.FileFromURL("https://elysium.fm" + mediaURL)
}

func (h *handler) inlineKeyboard(rows [][]entity.InlineButton) *telego.InlineKeyboardMarkup {
	if len(rows) == 0 {
		return nil
	}
	keyboard := make([][]telego.InlineKeyboardButton, 0, len(rows))
	for _, row := range rows {
		var buttons []telego.InlineKeyboardButton
		for _, button := range row {
			inline := telego.InlineKeyboardButton{Text: button.Text}
			if button.URL != "" {
				inline.URL = button.URL
			}
			if button.Data != "" {
				inline.CallbackData = button.Data
			}
			buttons = append(buttons, inline)
		}
		keyboard = append(keyboard, buttons)
	}
	return &telego.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func (h *handler) deleteMessage(chatID int64, messageID int) error {
	err := h.bot.DeleteMessage(context.Background(), &telego.DeleteMessageParams{
		ChatID:    tu.ID(chatID),
		MessageID: messageID,
	})
	return err
}

func (h *handler) answerCallback(callbackID, text string) {
	params := &telego.AnswerCallbackQueryParams{
		CallbackQueryID: callbackID,
	}
	if text != "" {
		params.Text = text
	}
	if err := h.bot.AnswerCallbackQuery(context.Background(), params); err != nil {
		h.logger.Error("answer callback query", "error", err)
	}
}

func defaultHourlySlots(open, close string) []string {
	startHour, _ := strconv.Atoi(strings.Split(open, ":")[0])
	endHour, _ := strconv.Atoi(strings.Split(close, ":")[0])
	if endHour <= startHour {
		endHour = startHour + 1
	}
	var slots []string
	for hour := startHour; hour < endHour; hour++ {
		slots = append(slots, fmt.Sprintf("%02d:00-%02d:00", hour, hour+1))
	}
	return slots
}

func (h *handler) userDisplayName(user *telego.User) string {
	if user == nil {
		return "Пользователь"
	}
	fullName := strings.TrimSpace(strings.TrimSpace(user.FirstName + " " + user.LastName))
	if fullName != "" {
		return fullName
	}
	if user.Username != "" {
		return "@" + user.Username
	}
	return fmt.Sprintf("user_%d", user.ID)
}
