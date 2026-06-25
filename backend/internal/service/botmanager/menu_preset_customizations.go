package botmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"revisitr/internal/entity"
)

type menuPresetCustomizations struct {
	Title          string                            `json:"title,omitempty"`
	Subtitle       string                            `json:"subtitle,omitempty"`
	CategoryOrder  []int                             `json:"category_order,omitempty"`
	Categories     []menuPresetCategoryCustomization `json:"categories,omitempty"`
	TabButtonStyle string                            `json:"tab_button_style,omitempty"`
	NavButtonStyle string                            `json:"nav_button_style,omitempty"`
	ListLayout     string                            `json:"list_layout,omitempty"`
	ListDensity    string                            `json:"list_density,omitempty"`
}

type menuPresetCategoryCustomization struct {
	CategoryID        int                          `json:"category_id"`
	Label             string                       `json:"label,omitempty"`
	IconImageURL      string                       `json:"icon_image_url,omitempty"`
	IconCustomEmojiID string                       `json:"icon_custom_emoji_id,omitempty"`
	Style             string                       `json:"style,omitempty"`
	EmojiOnly         bool                         `json:"emoji_only,omitempty"`
	ItemOrder         []int                        `json:"item_order,omitempty"`
	Items             []menuPresetItemCustomization `json:"items,omitempty"`
}

type menuPresetItemCustomization struct {
	ItemID int    `json:"item_id"`
	Label  string `json:"label,omitempty"`
	Hidden bool   `json:"hidden,omitempty"`
}

type menuCategoryPresentation struct {
	Category                entity.MenuCategory
	Label                   string
	Heading                 string
	ButtonText              string
	ButtonStyle             string
	ButtonIconCustomEmojiID string
}

const (
	menuListLayoutSummary  = "summary"
	menuListLayoutExpanded = "expanded"
	menuListDensityCompact = "compact"
	menuListDensityDetail  = "detailed"
)

func (h *handler) loadMenuPresetCustomizations(ctx context.Context) menuPresetCustomizations {
	if h.mgr.moduleSettingsRepo == nil {
		return menuPresetCustomizations{}
	}
	settings, err := h.mgr.moduleSettingsRepo.Get(ctx, h.info.ID, "menu")
	if err != nil || settings == nil || len(settings.Customizations) == 0 {
		return menuPresetCustomizations{}
	}

	var custom menuPresetCustomizations
	if err := json.Unmarshal(settings.Customizations, &custom); err != nil {
		h.logger.Error("decode menu preset customizations", "error", err, "bot_id", h.info.ID)
		return menuPresetCustomizations{}
	}
	custom.ListLayout = normalizeMenuListLayout(custom.ListLayout)
	custom.ListDensity = normalizeMenuListDensity(custom.ListDensity)
	return custom
}

func normalizeMenuListLayout(value string) string {
	switch strings.TrimSpace(value) {
	case menuListLayoutExpanded:
		return menuListLayoutExpanded
	default:
		return menuListLayoutSummary
	}
}

func normalizeMenuListDensity(value string) string {
	switch strings.TrimSpace(value) {
	case menuListDensityDetail:
		return menuListDensityDetail
	default:
		return menuListDensityCompact
	}
}

func (h *handler) presentMenuCategories(ctx context.Context, menu *entity.Menu, custom menuPresetCustomizations) []menuCategoryPresentation {
	if menu == nil || len(menu.Categories) == 0 {
		return nil
	}

	syncedEmojiByURL := h.syncedEmojiIDsByImageURL(ctx)

	customByID := make(map[int]menuPresetCategoryCustomization, len(custom.Categories))
	for _, item := range custom.Categories {
		if item.IconCustomEmojiID == "" && item.IconImageURL != "" {
			item.IconCustomEmojiID = syncedEmojiByURL[item.IconImageURL]
		}
		customByID[item.CategoryID] = item
	}

	orderIndex := make(map[int]int, len(custom.CategoryOrder))
	for i, id := range custom.CategoryOrder {
		orderIndex[id] = i
	}

	categories := append([]entity.MenuCategory(nil), menu.Categories...)
	sort.SliceStable(categories, func(i, j int) bool {
		leftIdx, leftOK := orderIndex[categories[i].ID]
		rightIdx, rightOK := orderIndex[categories[j].ID]
		switch {
		case leftOK && rightOK:
			return leftIdx < rightIdx
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return categories[i].SortOrder < categories[j].SortOrder
		}
	})

	presented := make([]menuCategoryPresentation, 0, len(categories))
	for _, category := range categories {
		override := customByID[category.ID]

		label := strings.TrimSpace(override.Label)
		displayLabel := label
		if displayLabel == "" {
			displayLabel = category.Name
		}

		buttonText := displayLabel
		if override.EmojiOnly {
			if override.IconCustomEmojiID != "" {
				buttonText = "⠀"
			} else if category.IconEmoji != nil && *category.IconEmoji != "" {
				buttonText = strings.TrimSpace(*category.IconEmoji)
			}
		} else if override.IconCustomEmojiID == "" && category.IconEmoji != nil && *category.IconEmoji != "" {
			buttonText = strings.TrimSpace(*category.IconEmoji + " " + displayLabel)
		}

		heading := displayLabel
		switch {
		case override.IconImageURL != "":
			heading = strings.TrimSpace("{{emoji:" + override.IconImageURL + "}} " + displayLabel)
		case category.IconEmoji != nil && *category.IconEmoji != "":
			heading = strings.TrimSpace(*category.IconEmoji + " " + displayLabel)
		}

		buttonStyle := strings.TrimSpace(override.Style)
		if buttonStyle == "" {
			buttonStyle = strings.TrimSpace(custom.TabButtonStyle)
		}

		category = applyItemCustomizations(category, override)

		presented = append(presented, menuCategoryPresentation{
			Category:                category,
			Label:                   displayLabel,
			Heading:                 heading,
			ButtonText:              buttonText,
			ButtonStyle:             buttonStyle,
			ButtonIconCustomEmojiID: strings.TrimSpace(override.IconCustomEmojiID),
		})
	}

	return presented
}

func (h *handler) syncedEmojiIDsByImageURL(ctx context.Context) map[string]string {
	if h == nil || h.mgr == nil || h.mgr.emojiRepo == nil {
		return nil
	}
	items, err := h.mgr.emojiRepo.GetSyncedItemsByOrgID(ctx, h.info.OrgID)
	if err != nil {
		h.logger.Error("load synced emoji ids", "error", err, "org_id", h.info.OrgID)
		return nil
	}

	byURL := make(map[string]string, len(items))
	for _, item := range items {
		if item.ImageURL == "" || item.TgCustomEmojiID == nil || *item.TgCustomEmojiID == "" {
			continue
		}
		byURL[item.ImageURL] = *item.TgCustomEmojiID
	}
	return byURL
}

func applyItemCustomizations(category entity.MenuCategory, override menuPresetCategoryCustomization) entity.MenuCategory {
	if len(override.ItemOrder) == 0 && len(override.Items) == 0 {
		return category
	}

	hiddenIDs := make(map[int]bool, len(override.Items))
	labelByID := make(map[int]string, len(override.Items))
	for _, item := range override.Items {
		if item.Hidden {
			hiddenIDs[item.ItemID] = true
		}
		if item.Label != "" {
			labelByID[item.ItemID] = item.Label
		}
	}

	items := make([]entity.MenuItem, 0, len(category.Items))
	for _, item := range category.Items {
		if hiddenIDs[item.ID] {
			continue
		}
		if label, ok := labelByID[item.ID]; ok {
			item.Name = label
		}
		items = append(items, item)
	}

	if len(override.ItemOrder) > 0 {
		orderIndex := make(map[int]int, len(override.ItemOrder))
		for i, id := range override.ItemOrder {
			orderIndex[id] = i
		}
		sort.SliceStable(items, func(i, j int) bool {
			leftIdx, leftOK := orderIndex[items[i].ID]
			rightIdx, rightOK := orderIndex[items[j].ID]
			switch {
			case leftOK && rightOK:
				return leftIdx < rightIdx
			case leftOK:
				return true
			case rightOK:
				return false
			default:
				return items[i].SortOrder < items[j].SortOrder
			}
		})
	}

	category.Items = items
	return category
}

func menuCategoryItemsTextWithDensity(category entity.MenuCategory, density string) string {
	lines := make([]string, 0, len(category.Items))
	for _, item := range category.Items {
		if !item.IsAvailable {
			continue
		}

		line := item.Name + " — " + strings.TrimSpace(formatMenuPrice(item.Price))
		if strings.TrimSpace(density) != menuListDensityCompact && item.Weight != nil && *item.Weight != "" {
			line += " • " + strings.TrimSpace(*item.Weight)
		}
		lines = append(lines, line)

		if strings.TrimSpace(density) != menuListDensityCompact && item.Description != nil && *item.Description != "" {
			lines = append(lines, strings.TrimSpace(*item.Description))
		}
	}

	if len(lines) == 0 {
		return "Сейчас в этой категории нет доступных позиций."
	}
	return strings.Join(lines, "\n")
}

func formatMenuPrice(price float64) string {
	return fmt.Sprintf("%.0f ₽", price)
}

func menuSections(sections ...string) string {
	var cleaned []string
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section != "" {
			cleaned = append(cleaned, section)
		}
	}
	return strings.Join(cleaned, "\n\n")
}

func (h *handler) buildMenuHeader(menu *entity.Menu, custom menuPresetCustomizations) string {
	title := strings.TrimSpace(custom.Title)
	if title == "" && menu != nil {
		title = strings.TrimSpace(menu.Name)
	}

	subtitle := strings.TrimSpace(custom.Subtitle)
	return menuSections(title, subtitle)
}
