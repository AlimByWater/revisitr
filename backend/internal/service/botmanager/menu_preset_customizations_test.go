package botmanager

import (
	"testing"

	"revisitr/internal/entity"
)

func TestPresentMenuCategoriesAppliesOrderAndOverrides(t *testing.T) {
	h := &handler{}
	icon := "🍝"
	menu := &entity.Menu{
		Categories: []entity.MenuCategory{
			{ID: 1, Name: "Starters", SortOrder: 0},
			{ID: 2, Name: "Pasta", SortOrder: 1, IconEmoji: &icon},
		},
	}

	custom := menuPresetCustomizations{
		CategoryOrder:  []int{2, 1},
		TabButtonStyle: "primary",
		Categories: []menuPresetCategoryCustomization{
			{
				CategoryID:        2,
				Label:             "House Pasta",
				IconImageURL:      "/emoji/pasta.png",
				IconCustomEmojiID: "ce_123",
				Style:             "success",
			},
		},
	}

	got := h.presentMenuCategories(nil, menu, custom)
	if len(got) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(got))
	}
	if got[0].Category.ID != 2 {
		t.Fatalf("expected reordered category first, got id=%d", got[0].Category.ID)
	}
	if got[0].Label != "House Pasta" {
		t.Fatalf("expected overridden label, got %q", got[0].Label)
	}
	if got[0].Heading != "{{emoji:/emoji/pasta.png}} House Pasta" {
		t.Fatalf("expected emoji marker heading, got %q", got[0].Heading)
	}
	if got[0].ButtonIconCustomEmojiID != "ce_123" {
		t.Fatalf("expected custom emoji id, got %q", got[0].ButtonIconCustomEmojiID)
	}
	if got[0].ButtonStyle != "success" {
		t.Fatalf("expected override style, got %q", got[0].ButtonStyle)
	}
	if got[1].ButtonStyle != "primary" {
		t.Fatalf("expected fallback tab style, got %q", got[1].ButtonStyle)
	}
}

func TestBuildCategoryTabRowsMarksActiveCategory(t *testing.T) {
	rows := buildCategoryTabRows([]menuCategoryPresentation{
		{
			Category:                entity.MenuCategory{ID: 10},
			ButtonText:              "Desserts",
			ButtonStyle:             "primary",
			ButtonIconCustomEmojiID: "emoji_1",
		},
		{
			Category:   entity.MenuCategory{ID: 20},
			ButtonText: "Drinks",
		},
	}, 10)

	if len(rows) != 1 || len(rows[0]) != 2 {
		t.Fatalf("unexpected row layout: %#v", rows)
	}
	if rows[0][0].Text != "Desserts" {
		t.Fatalf("expected tab label, got %q", rows[0][0].Text)
	}
	if rows[0][0].Style != "success" {
		t.Fatalf("expected active success style, got %q", rows[0][0].Style)
	}
	if rows[0][0].IconCustomEmojiID != "emoji_1" {
		t.Fatalf("expected custom emoji id preserved, got %q", rows[0][0].IconCustomEmojiID)
	}
	if rows[0][0].Data != callbackMenuTabPref+"10" {
		t.Fatalf("unexpected callback data %q", rows[0][0].Data)
	}
}

func TestPresentMenuCategoriesSupportsEmojiOnlyTabs(t *testing.T) {
	h := &handler{}
	icon := "🍰"
	menu := &entity.Menu{
		Categories: []entity.MenuCategory{
			{ID: 1, Name: "Desserts", SortOrder: 0, IconEmoji: &icon},
		},
	}

	got := h.presentMenuCategories(nil, menu, menuPresetCustomizations{
		Categories: []menuPresetCategoryCustomization{{
			CategoryID: 1,
			EmojiOnly:  true,
		}},
	})
	if len(got) != 1 {
		t.Fatalf("expected one category, got %d", len(got))
	}
	if got[0].ButtonText != "🍰" {
		t.Fatalf("expected emoji-only button text, got %q", got[0].ButtonText)
	}
}
