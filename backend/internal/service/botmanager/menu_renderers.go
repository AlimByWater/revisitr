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

const menuTabsPerRow = 4
const menuItemButtonsPerRow = 1

func (h *handler) renderMenuTabs(ctx context.Context, chatID int64, selectedCategoryID int) {
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

	selected := categories[0]
	for _, category := range categories {
		if category.Category.ID == selectedCategoryID {
			selected = category
			break
		}
	}

	part := h.menuBasePart(menu)
	part.Text = truncateMenuText(menuSections(
		h.buildMenuTabsHeader(menu, custom),
		renderMenuASCIIBlock(selected.Category.Items),
	))
	part = ensureMenuTextPart(part)

	content := entity.MessageContent{
		Parts: []entity.MessagePart{part},
		Buttons: append(
			buildCategoryTabRows(categories, selected.Category.ID),
			buildCategoryItemRows(selected.Category.ID, selected.Category.Items)...,
		),
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu_tabs"
	state.MenuCategoryID = selected.Category.ID
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

func buildCategoryTabRows(categories []menuCategoryPresentation, activeCategoryID int) [][]entity.InlineButton {
	rows := make([][]entity.InlineButton, 0, (len(categories)+menuTabsPerRow-1)/menuTabsPerRow)
	current := make([]entity.InlineButton, 0, menuTabsPerRow)

	for _, category := range categories {
		text := category.ButtonText
		style := category.ButtonStyle
		if category.Category.ID == activeCategoryID {
			style = "success"
		}

		current = append(current, entity.InlineButton{
			Text:              text,
			Data:              callbackMenuTabPref + strconv.Itoa(category.Category.ID),
			Style:             style,
			IconCustomEmojiID: category.ButtonIconCustomEmojiID,
		})

		if len(current) == menuTabsPerRow {
			rows = append(rows, current)
			current = make([]entity.InlineButton, 0, menuTabsPerRow)
		}
	}

	if len(current) > 0 {
		rows = append(rows, current)
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

	sections := []string{h.buildMenuHeader(menu, custom)}
	for _, category := range categories {
		sections = append(sections, menuSections(category.Heading, menuCategoryItemsText(category.Category)))
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
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, h.carouselContent(items, index, custom))
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

	text := menuSections(
		item.CategoryHeading,
		item.Name+" — "+formatMenuPrice(item.Price),
		strings.TrimSpace(valueOrEmpty(item.Weight)),
		strings.TrimSpace(valueOrEmpty(item.Description)),
	)
	text = truncateMenuText(text)

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
			Text:      text,
			ParseMode: "HTML",
		}}
	} else {
		content.Parts = []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      text,
			ParseMode: "HTML",
		}}
	}

	return content
}

// sendCarouselItem sends a carousel card and returns the message ID.
func (h *handler) sendCarouselItem(ctx context.Context, chatID int64, items []carouselItem, index int, custom menuPresetCustomizations) int {
	return h.sendContentMessage(ctx, chatID, h.carouselContent(items, index, custom))
}

func truncateMenuText(text string) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= 4000 {
		return string(runes)
	}
	return string(runes[:4000]) + "\n\n... ещё позиции"
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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

	_ = h.updateMessageByID(ctx, chatID, query.Message.GetMessageID(), content)
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
	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		return entity.MessageContent{}, false
	}

	custom := h.loadMenuPresetCustomizations(ctx)
	categories := h.presentMenuCategories(ctx, menu, custom)
	category, index, items, ok := findCategoryItem(categories, categoryID, itemID)
	if !ok {
		return entity.MessageContent{}, false
	}

	item := items[index]
	text := menuSections(
		category.Heading,
		html.EscapeString(item.Name)+" — "+formatMenuPrice(item.Price),
		html.EscapeString(strings.TrimSpace(valueOrEmpty(item.Weight))),
		html.EscapeString(strings.TrimSpace(valueOrEmpty(item.Description))),
	)

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
			Text:      truncateMenuText(text),
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
