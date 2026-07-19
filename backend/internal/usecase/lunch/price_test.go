package lunch

import (
	"testing"

	"revisitr/internal/entity"
)

func TestCalculateTotalFixed(t *testing.T) {
	total, err := CalculateTotal(LunchPriceInput{
		PriceMode: entity.LunchPriceFixed,
		BasePrice: 350,
		Items: []LunchPriceItem{
			{MenuItemPrice: 200, Surcharge: 50},
			{MenuItemPrice: 300, Surcharge: 0},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 350 {
		t.Errorf("fixed: got %v, want 350 (item prices and surcharges must be ignored)", total)
	}
}

func TestCalculateTotalSumOfItems(t *testing.T) {
	total, err := CalculateTotal(LunchPriceInput{
		PriceMode: entity.LunchPriceSumOfItems,
		BasePrice: 999, // must be ignored
		Items: []LunchPriceItem{
			{MenuItemPrice: 150.50, Surcharge: 100},
			{MenuItemPrice: 249.50, Surcharge: 200},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 400 {
		t.Errorf("sum_of_items: got %v, want 400 (base price and surcharges must be ignored)", total)
	}
}

func TestCalculateTotalBasePlusSurcharge(t *testing.T) {
	total, err := CalculateTotal(LunchPriceInput{
		PriceMode: entity.LunchPriceBasePlusSurcharge,
		BasePrice: 300,
		Items: []LunchPriceItem{
			{MenuItemPrice: 500, Surcharge: 200}, // стейк +200
			{MenuItemPrice: 100, Surcharge: 0},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 500 {
		t.Errorf("base_plus_surcharge: got %v, want 500 (menu prices must be ignored)", total)
	}
}

func TestCalculateTotalEmptyItems(t *testing.T) {
	total, err := CalculateTotal(LunchPriceInput{
		PriceMode: entity.LunchPriceSumOfItems,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("empty sum: got %v, want 0", total)
	}
}

func TestCalculateTotalUnknownMode(t *testing.T) {
	if _, err := CalculateTotal(LunchPriceInput{PriceMode: "percent"}); err == nil {
		t.Error("expected error for unknown price mode")
	}
}

func minTotalFixtures() (entity.LunchFormat, []entity.LunchCourse) {
	courses := []entity.LunchCourse{
		{
			ID: 1,
			Items: []entity.LunchCourseItem{
				{Surcharge: 0, MenuItem: &entity.MenuItem{Price: 180}},
				{Surcharge: 50, MenuItem: &entity.MenuItem{Price: 120}},
			},
		},
		{
			ID: 2,
			Items: []entity.LunchCourseItem{
				{Surcharge: 200, MenuItem: &entity.MenuItem{Price: 500}},
				{Surcharge: 20, MenuItem: &entity.MenuItem{Price: 250}},
			},
		},
	}
	format := entity.LunchFormat{BasePrice: 300, CourseIDs: []int{1, 2}}
	return format, courses
}

func TestMinTotal(t *testing.T) {
	format, courses := minTotalFixtures()

	format.PriceMode = entity.LunchPriceFixed
	if got := MinTotal(format, courses); got != 300 {
		t.Errorf("fixed: got %v, want 300", got)
	}

	format.PriceMode = entity.LunchPriceSumOfItems
	if got := MinTotal(format, courses); got != 370 { // 120 + 250
		t.Errorf("sum_of_items: got %v, want 370", got)
	}

	format.PriceMode = entity.LunchPriceBasePlusSurcharge
	if got := MinTotal(format, courses); got != 320 { // 300 + 0 + 20
		t.Errorf("base_plus_surcharge: got %v, want 320", got)
	}
}
