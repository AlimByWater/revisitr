package botmanager

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
)

// renderMenuList renders all menu items as a flat grouped list in one message.
func (h *handler) renderMenuList(ctx context.Context, chatID int64) {
	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>%s</b>\n\n", menu.Name))
	for _, cat := range menu.Categories {
		label := cat.Name
		if cat.IconEmoji != nil && *cat.IconEmoji != "" {
			label = *cat.IconEmoji + " " + label
		}
		sb.WriteString(fmt.Sprintf("<b>%s</b>\n", label))
		for _, item := range cat.Items {
			if !item.IsAvailable {
				continue
			}
			price := fmt.Sprintf("%.0f ₽", item.Price)
			sb.WriteString(fmt.Sprintf("  • %s — %s\n", item.Name, price))
			if item.Description != nil && *item.Description != "" {
				sb.WriteString(fmt.Sprintf("    <i>%s</i>\n", *item.Description))
			}
		}
		sb.WriteString("\n")
	}

	// Rune-safe truncation at Telegram limit
	text := sb.String()
	runes := []rune(text)
	if len(runes) > 4000 {
		text = string(runes[:4000]) + "\n\n... ещё позиции"
	}

	_, _ = h.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: chatID},
		Text:      text,
		ParseMode: telego.ModeHTML,
	})
}

// carouselItem holds a menu item with its category name for carousel display.
type carouselItem struct {
	entity.MenuItem
	CategoryName string
}

// renderMenuCarousel shows the first menu item as a card with left/right navigation.
func (h *handler) renderMenuCarousel(ctx context.Context, chatID int64) {
	items := h.flattenMenuItems(ctx, chatID)
	if items == nil {
		h.sendText(chatID, h.menuUnavailableMessage())
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.Flow = "menu_carousel"
	state.CarouselIndex = 0
	state.CarouselTotal = len(items)
	state.FlowMessageID = h.sendCarouselItem(ctx, chatID, items, 0)
	_ = h.saveFlowState(ctx, chatID, state)
}

// handleMenuCarouselCallback handles ←/→ navigation in the carousel.
func (h *handler) handleMenuCarouselCallback(ctx context.Context, chatID int64, value string) {
	if value == "noop" {
		return
	}
	index, err := strconv.Atoi(value)
	if err != nil {
		return
	}

	items := h.flattenMenuItems(ctx, chatID)
	if items == nil || index < 0 || index >= len(items) {
		return
	}

	state := h.currentFlowState(ctx, chatID)
	state.CarouselIndex = index
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, h.carouselContent(items, index))
	_ = h.saveFlowState(ctx, chatID, state)
}

// flattenMenuItems loads and flattens all available menu items across categories.
func (h *handler) flattenMenuItems(ctx context.Context, chatID int64) []carouselItem {
	menu, _, ok := h.resolveMenuForChat(ctx, chatID)
	if !ok || menu == nil {
		return nil
	}

	var items []carouselItem
	for _, cat := range menu.Categories {
		catName := cat.Name
		if cat.IconEmoji != nil && *cat.IconEmoji != "" {
			catName = *cat.IconEmoji + " " + catName
		}
		for _, item := range cat.Items {
			if !item.IsAvailable {
				continue
			}
			items = append(items, carouselItem{MenuItem: item, CategoryName: catName})
		}
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

// carouselContent builds a MessageContent for a single carousel item.
func (h *handler) carouselContent(items []carouselItem, index int) entity.MessageContent {
	item := items[index]
	total := len(items)

	var descLine string
	if item.Description != nil && *item.Description != "" {
		descLine = "\n" + *item.Description
	}

	text := fmt.Sprintf("<b>%s</b>\n%s — %.0f ₽%s",
		item.CategoryName, item.Name, item.Price, descLine)

	// Navigation buttons
	var nav []entity.InlineButton
	if index > 0 {
		nav = append(nav, entity.InlineButton{Text: "◀️", Data: callbackMenuCarouselPref + strconv.Itoa(index-1)})
	}
	nav = append(nav, entity.InlineButton{
		Text: fmt.Sprintf("%d/%d", index+1, total),
		Data: callbackMenuCarouselPref + "noop",
	})
	if index < total-1 {
		nav = append(nav, entity.InlineButton{Text: "▶️", Data: callbackMenuCarouselPref + strconv.Itoa(index+1)})
	}

	content := entity.MessageContent{
		Buttons: [][]entity.InlineButton{nav},
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
func (h *handler) sendCarouselItem(ctx context.Context, chatID int64, items []carouselItem, index int) int {
	content := h.carouselContent(items, index)

	// Use sendContentMessage which handles photo + text + inline buttons correctly
	return h.sendContentMessage(ctx, chatID, content)
}
