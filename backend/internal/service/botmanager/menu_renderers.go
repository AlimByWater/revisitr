package botmanager

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
)

const menuItemButtonsPerRow = 1

func (h *handler) renderMenuTabs(ctx context.Context, chatID int64, selectedCategoryID int) {
	h.logger.Info("renderMenuTabs called", "chat_id", chatID, "selected_category_id", selectedCategoryID)

	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.logger.Warn("renderMenuTabs: no menu resolved for chat", "chat_id", chatID)
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}
	h.logger.Info("renderMenuTabs: menu resolved", "chat_id", chatID, "menu_id", menu.ID, "menu_name", menu.Name)

	custom := h.loadMenuPresetCustomizations(ctx)
	categories := h.presentMenuCategories(ctx, menu, custom)
	h.logger.Info("renderMenuTabs: categories presented", "chat_id", chatID, "category_count", len(categories))
	if len(categories) == 0 {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	var selected *menuCategoryPresentation
	for i := range categories {
		if categories[i].Category.ID == selectedCategoryID {
			selected = &categories[i]
			break
		}
	}

	activeCategoryID := 0
	var itemsText string
	var itemRows [][]entity.InlineButton
	if selected != nil {
		h.logger.Info("renderMenuTabs: selected category", "chat_id", chatID, "category_id", selected.Category.ID, "category_name", selected.Category.Name, "item_count", len(selected.Category.Items))
		activeCategoryID = selected.Category.ID
		itemsText = renderMenuASCIIBlock(selected.Category.Items)
		itemRows = buildCategoryItemRows(selected.Category.ID, selected.Category.Items)
	}

	part := h.menuBasePart(menu)
	part.Text = truncateMenuText(menuSections(
		h.buildMenuTabsHeader(menu, custom),
		itemsText,
	))
	part = ensureMenuTextPart(part)

	var buttons [][]entity.InlineButton
	if selected != nil {
		buttons = append(buttons, itemRows...)
		buttons = append(buttons, []entity.InlineButton{{Text: "◀ Назад к категориям", Data: callbackMenuRoot}})
	} else {
		buttons = buildCategoryTabRows(categories, activeCategoryID)
	}

	content := entity.MessageContent{
		Parts:   []entity.MessagePart{part},
		Buttons: buttons,
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu_tabs"
	state.MenuCategoryID = activeCategoryID
	state.MenuPage = 0
	state.FlowMessageID = h.updateFlowMessage(ctx, chatID, state, content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) menuBasePart(menu *entity.Menu) entity.MessagePart {
	intro := h.menuIntroContent(menu)
	if len(intro.Parts) > 0 {
		part := intro.Parts[0]
		if part.Type == "" {
			part.Type = entity.PartText
		}
		return part
	}
	return entity.MessagePart{Type: entity.PartText}
}

// Telegram splits a row's button widths evenly across however many buttons
// share that row, and does so inconsistently between clients (desktop, Android,
// iOS) once labels of different lengths mix. One category per row sidesteps
// that entirely: a row with a single button can't be split unevenly.
func buildCategoryTabRows(categories []menuCategoryPresentation, activeCategoryID int) [][]entity.InlineButton {
	rows := make([][]entity.InlineButton, 0, len(categories))
	for _, category := range categories {
		style := category.ButtonStyle
		if category.Category.ID == activeCategoryID {
			style = "success"
		}

		rows = append(rows, []entity.InlineButton{{
			Text:              category.ButtonText,
			Data:              callbackMenuTabPref + strconv.Itoa(category.Category.ID),
			Style:             style,
			IconCustomEmojiID: category.ButtonIconCustomEmojiID,
		}})
	}
	return rows
}

func buildCategoryItemRows(categoryID int, items []entity.MenuItem) [][]entity.InlineButton {
	rows := make([][]entity.InlineButton, 0, len(items))
	current := make([]entity.InlineButton, 0, menuItemButtonsPerRow)

	for _, item := range items {
		if !item.IsAvailable {
			continue
		}

		current = append(current, entity.InlineButton{
			Text: item.Name,
			Data: callbackMenuCardPref + strconv.Itoa(item.ID) + ":" + strconv.Itoa(categoryID),
		})

		if len(current) == menuItemButtonsPerRow {
			rows = append(rows, current)
			current = make([]entity.InlineButton, 0, menuItemButtonsPerRow)
		}
	}

	if len(current) > 0 {
		rows = append(rows, current)
	}

	return rows
}

func buildMenuListCategoryRows(categories []menuCategoryPresentation) [][]entity.InlineButton {
	rows := make([][]entity.InlineButton, 0, len(categories))
	for _, category := range categories {
		rows = append(rows, []entity.InlineButton{{
			Text:              category.ButtonText,
			Data:              callbackMenuCategoryPref + strconv.Itoa(category.Category.ID) + ":0",
			Style:             category.ButtonStyle,
			IconCustomEmojiID: category.ButtonIconCustomEmojiID,
		}})
	}
	return rows
}

// renderMenuList renders all menu items as a flat grouped list in one message.
func (h *handler) renderMenuList(ctx context.Context, chatID int64) {
	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	custom := h.loadMenuPresetCustomizations(ctx)
	categories := h.presentMenuCategories(ctx, menu, custom)
	if len(categories) == 0 {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	layout := normalizeMenuListLayout(custom.ListLayout)
	density := normalizeMenuListDensity(custom.ListDensity)

	sections := []string{h.buildMenuHeader(menu, custom)}
	if layout == menuListLayoutExpanded {
		for _, category := range categories {
			sections = append(sections, menuSections(category.Heading, menuCategoryItemsTextWithDensity(category.Category, density)))
		}
	} else {
		sections = append(sections, "Выберите категорию")
	}

	part := h.menuBasePart(menu)
	part.Text = truncateMenuText(menuSections(sections...))
	part = ensureMenuTextPart(part)

	h.sendContentMessage(ctx, chatID, entity.MessageContent{
		Parts:   []entity.MessagePart{part},
		Buttons: buildMenuListCategoryRows(categories),
	})
}

type carouselItem struct {
	entity.MenuItem
	CategoryHeading string
}

// renderMenuCarousel shows the first menu item as a card with left/right navigation.
func (h *handler) renderMenuCarousel(ctx context.Context, chatID int64) {
	items, custom := h.flattenMenuItems(ctx, chatID)
	if items == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu_carousel"
	state.CarouselIndex = 0
	state.CarouselTotal = len(items)
	state.FlowMessageID = h.sendCarouselItem(ctx, chatID, items, 0, custom)
	_ = h.saveFlowState(ctx, chatID, state)
}

// handleMenuCarouselCallback handles <-/-> navigation in the carousel.
func (h *handler) handleMenuCarouselCallback(ctx context.Context, chatID int64, value string) {
	if value == "noop" {
		return
	}
	index, err := strconv.Atoi(value)
	if err != nil {
		return
	}

	items, custom := h.flattenMenuItems(ctx, chatID)
	if items == nil || index < 0 || index >= len(items) {
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.CarouselIndex = index
	state.FlowMessageID = h.updateFlowMessage(ctx, chatID, state, h.carouselContent(items, index, custom))
	_ = h.saveFlowState(ctx, chatID, state)
}

// flattenMenuItems loads and flattens all available menu items across categories.
func (h *handler) flattenMenuItems(ctx context.Context, chatID int64) ([]carouselItem, menuPresetCustomizations) {
	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		return nil, menuPresetCustomizations{}
	}

	custom := h.loadMenuPresetCustomizations(ctx)
	categories := h.presentMenuCategories(ctx, menu, custom)
	if len(categories) == 0 {
		return nil, custom
	}

	var items []carouselItem
	for _, category := range categories {
		for _, item := range category.Category.Items {
			if !item.IsAvailable {
				continue
			}
			items = append(items, carouselItem{
				MenuItem:        item,
				CategoryHeading: category.Heading,
			})
		}
	}
	if len(items) == 0 {
		return nil, custom
	}
	return items, custom
}

// carouselContent builds a MessageContent for a single carousel item.
func (h *handler) carouselContent(items []carouselItem, index int, custom menuPresetCustomizations) entity.MessageContent {
	item := items[index]
	total := len(items)

	rawText := formatMenuItemCardText(item.CategoryHeading, item.MenuItem)

	var nav []entity.InlineButton
	if index > 0 {
		nav = append(nav, entity.InlineButton{
			Text:  "←",
			Data:  callbackMenuCarouselPref + strconv.Itoa(index-1),
			Style: custom.NavButtonStyle,
		})
	}
	nav = append(nav, entity.InlineButton{
		Text:  fmt.Sprintf("%d/%d", index+1, total),
		Data:  callbackMenuCarouselPref + "noop",
		Style: custom.NavButtonStyle,
	})
	if index < total-1 {
		nav = append(nav, entity.InlineButton{
			Text:  "→",
			Data:  callbackMenuCarouselPref + strconv.Itoa(index+1),
			Style: custom.NavButtonStyle,
		})
	}

	content := entity.MessageContent{
		Buttons: [][]entity.InlineButton{
			nav,
			{{Text: "Назад к меню", Data: callbackMenuRoot}},
		},
	}

	if item.ImageURL != nil && *item.ImageURL != "" {
		content.Parts = []entity.MessagePart{{
			Type:      entity.PartPhoto,
			MediaURL:  *item.ImageURL,
			Text:      truncateMenuCaption(rawText),
			ParseMode: "HTML",
		}}
	} else {
		content.Parts = []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      truncateMenuText(rawText),
			ParseMode: "HTML",
		}}
	}

	return content
}

// sendCarouselItem sends a carousel card and returns the message ID.
func (h *handler) sendCarouselItem(ctx context.Context, chatID int64, items []carouselItem, index int, custom menuPresetCustomizations) int {
	return h.sendContentMessage(ctx, chatID, h.carouselContent(items, index, custom))
}

func formatMenuItemCardText(heading string, item entity.MenuItem) string {
	lines := []string{
		"/// " + html.EscapeString(heading),
		"<b>" + html.EscapeString(item.Name) + " — " + formatMenuPrice(item.Price) + "</b>",
	}
	if item.Weight != nil && strings.TrimSpace(*item.Weight) != "" {
		lines = append(lines, "<i>"+html.EscapeString(strings.TrimSpace(*item.Weight))+"</i>")
	}
	text := strings.Join(lines, "\n")
	if item.Description != nil && strings.TrimSpace(*item.Description) != "" {
		text += "\n\n" + html.EscapeString(strings.TrimSpace(*item.Description))
	}
	return text
}

const menuTextMaxLen = 4000

// menuCaptionMaxLen is Telegram's hard limit for photo captions (1024 chars),
// unlike the 4096-char limit for plain text messages.
const menuCaptionMaxLen = 1024

func truncateMenuText(text string) string {
	return truncateMenuTextTo(text, menuTextMaxLen)
}

func truncateMenuCaption(text string) string {
	return truncateMenuTextTo(text, menuCaptionMaxLen)
}

func truncateMenuTextTo(text string, max int) string {
	const suffix = "\n\n... ещё позиции"
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	cut := max - len([]rune(suffix))
	if cut < 0 {
		cut = 0
	}
	return string(runes[:cut]) + suffix
}

func renderMenuASCIIBlock(items []entity.MenuItem) string {
	lines := []string{
		"╔═.·:·. ﹏﹏𓂁🦈𓂁﹏﹏.·:·.═╗",
		"║",
	}

	for _, item := range items {
		if !item.IsAvailable {
			continue
		}
		lines = append(lines, "║ ⟼ "+item.Name)
	}

	lines = append(lines,
		"║",
		"╚═.·:·. ﹏﹏𓂁🪸𓂁﹏﹏.·:·.═╝",
	)

	return strings.Join(lines, "\n")
}

func (h *handler) buildMenuTabsHeader(menu *entity.Menu, custom menuPresetCustomizations) string {
	title := strings.TrimSpace(custom.Title)
	if title == "" && menu != nil {
		title = strings.TrimSpace(menu.Name)
	}

	subtitle := strings.TrimSpace(custom.Subtitle)
	return menuSections(title, subtitle)
}

func ensureMenuTextPart(part entity.MessagePart) entity.MessagePart {
	if part.Type == entity.PartText {
		return part
	}
	if len([]rune(part.Text)) <= 900 {
		return part
	}

	return entity.MessagePart{
		Type:      entity.PartText,
		Text:      part.Text,
		ParseMode: part.ParseMode,
	}
}

func (h *handler) handleMenuCardCallback(ctx context.Context, query *telego.CallbackQuery, value string) {
	if query == nil || query.Message == nil {
		h.logger.Warn("handleMenuCardCallback: nil query or message")
		return
	}
	h.logger.Info("handleMenuCardCallback", "chat_id", query.Message.GetChat().ID, "value", value)
	itemID, categoryID, ok := parseMenuCardValue(value)
	if !ok {
		h.logger.Warn("handleMenuCardCallback: failed to parse value", "value", value)
		return
	}
	h.logger.Info("handleMenuCardCallback: parsed", "item_id", itemID, "category_id", categoryID)

	chatID := query.Message.GetChat().ID
	content, ok := h.menuItemCardContent(ctx, chatID, itemID, categoryID)
	if !ok {
		h.logger.Warn("handleMenuCardCallback: menuItemCardContent returned false", "chat_id", chatID, "item_id", itemID, "category_id", categoryID)
		return
	}
	h.logger.Info("handleMenuCardCallback: content ready, message_id", "chat_id", chatID, "msg_id", query.Message.GetMessageID())

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu"
	state.FlowMessageID = h.updateMessageByID(ctx, chatID, query.Message.GetMessageID(), content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleMenuCardNavigationCallback(ctx context.Context, query *telego.CallbackQuery, value string) {
	if query == nil || query.Message == nil {
		return
	}
	itemID, categoryID, ok := parseMenuCardValue(value)
	if !ok {
		return
	}

	chatID := query.Message.GetChat().ID
	content, ok := h.menuItemCardContent(ctx, chatID, itemID, categoryID)
	if !ok {
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu"
	state.FlowMessageID = h.updateMessageByID(ctx, chatID, query.Message.GetMessageID(), content)
	_ = h.saveFlowState(ctx, chatID, state)
}

func parseMenuCardValue(value string) (itemID int, categoryID int, ok bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}

	itemID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	categoryID, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}

	return itemID, categoryID, true
}

func (h *handler) menuItemCardContent(ctx context.Context, chatID int64, itemID int, categoryID int) (entity.MessageContent, bool) {
	h.logger.Info("menuItemCardContent", "chat_id", chatID, "item_id", itemID, "category_id", categoryID)

	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.logger.Warn("menuItemCardContent: no menu resolved", "chat_id", chatID)
		return entity.MessageContent{}, false
	}
	h.logger.Info("menuItemCardContent: menu resolved", "chat_id", chatID, "menu_id", menu.ID)

	custom := h.loadMenuPresetCustomizations(ctx)
	categories := h.presentMenuCategories(ctx, menu, custom)
	h.logger.Info("menuItemCardContent: categories presented", "chat_id", chatID, "category_count", len(categories))

	category, index, items, ok := findCategoryItem(categories, categoryID, itemID)
	if !ok {
		h.logger.Warn("menuItemCardContent: findCategoryItem failed", "chat_id", chatID, "category_id", categoryID, "item_id", itemID)
		return entity.MessageContent{}, false
	}
	h.logger.Info("menuItemCardContent: item found", "chat_id", chatID, "category_name", category.Category.Name, "item_index", index, "total_items", len(items))

	item := items[index]
	text := formatMenuItemCardText(category.Heading, item)

	var rows [][]entity.InlineButton
	var nav []entity.InlineButton
	if index > 0 {
		nav = append(nav, entity.InlineButton{
			Text:  "←",
			Data:  callbackMenuCardNavPref + strconv.Itoa(items[index-1].ID) + ":" + strconv.Itoa(categoryID),
			Style: custom.NavButtonStyle,
		})
	}
	nav = append(nav, entity.InlineButton{
		Text:  fmt.Sprintf("%d/%d", index+1, len(items)),
		Data:  callbackMenuNoop,
		Style: custom.NavButtonStyle,
	})
	if index < len(items)-1 {
		nav = append(nav, entity.InlineButton{
			Text:  "→",
			Data:  callbackMenuCardNavPref + strconv.Itoa(items[index+1].ID) + ":" + strconv.Itoa(categoryID),
			Style: custom.NavButtonStyle,
		})
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}
	rows = append(rows, []entity.InlineButton{{
		Text: "Назад к категории",
		Data: callbackMenuTabPref + strconv.Itoa(categoryID),
	}})
	rows = append(rows, []entity.InlineButton{{
		Text: "✕ Закрыть",
		Data: callbackMenuCardClose,
	}})

	content := entity.MessageContent{
		Buttons: rows,
	}

	if item.ImageURL != nil && *item.ImageURL != "" {
		content.Parts = []entity.MessagePart{{
			Type:      entity.PartPhoto,
			MediaURL:  *item.ImageURL,
			Text:      truncateMenuCaption(text),
			ParseMode: "HTML",
		}}
	} else {
		content.Parts = []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      truncateMenuText(text),
			ParseMode: "HTML",
		}}
	}

	return content, true
}

func findCategoryItem(categories []menuCategoryPresentation, categoryID int, itemID int) (menuCategoryPresentation, int, []entity.MenuItem, bool) {
	for _, category := range categories {
		if category.Category.ID != categoryID {
			continue
		}

		var items []entity.MenuItem
		selectedIndex := -1
		for _, item := range category.Category.Items {
			if !item.IsAvailable {
				continue
			}
			if item.ID == itemID {
				selectedIndex = len(items)
			}
			items = append(items, item)
		}

		if selectedIndex >= 0 {
			return category, selectedIndex, items, true
		}
	}

	return menuCategoryPresentation{}, 0, nil, false
}
