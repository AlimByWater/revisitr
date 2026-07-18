package botmanager

import (
	"testing"

	"revisitr/internal/entity"
)

func TestBuildLunchOrderSnapshots(t *testing.T) {
	program := lunchTestProgram()
	format := lunchFormatByID(program, 5) // fixed 350
	courses := lunchFormatCourses(program, format)

	order, err := buildLunchOrder(2, 42, format, courses, map[int]int{10: 100, 20: 200}, "7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.BotID != 2 || order.BotClientID != 42 || order.TableNum != "7" {
		t.Errorf("order identity mismatch: %+v", order)
	}
	if order.Status != entity.OrderStatusNew {
		t.Errorf("new order must have status new, got %q", order.Status)
	}
	if order.Source != entity.OrderSourceLunch {
		t.Errorf("lunch flow must stamp source lunch, got %q", order.Source)
	}
	if order.FormatID == nil || *order.FormatID != 5 || order.FormatName != "Первое + Второе" {
		t.Errorf("format snapshot mismatch: %+v", order)
	}
	if order.TotalPrice != 350 {
		t.Errorf("fixed total: got %v, want 350", order.TotalPrice)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}
	first := order.Items[0]
	if first.CourseTitle != "Первое" || first.ItemName != "Борщ" || first.Price != 180 {
		t.Errorf("item snapshot mismatch: %+v", first)
	}
	second := order.Items[1]
	if second.ItemName != "Стейк" || second.Surcharge != 200 {
		t.Errorf("surcharge snapshot mismatch: %+v", second)
	}
}

func TestBuildLunchOrderMissingSelection(t *testing.T) {
	program := lunchTestProgram()
	format := lunchFormatByID(program, 5)
	courses := lunchFormatCourses(program, format)

	if _, err := buildLunchOrder(2, 42, format, courses, map[int]int{10: 100}, "7"); err == nil {
		t.Error("expected error when a course has no selection")
	}
}

func TestBuildLunchOrderGoneItem(t *testing.T) {
	program := lunchTestProgram()
	format := lunchFormatByID(program, 5)
	courses := lunchFormatCourses(program, format)

	// Item 999 was never (or is no longer) part of the course — stop-list race.
	if _, err := buildLunchOrder(2, 42, format, courses, map[int]int{10: 999, 20: 200}, "7"); err == nil {
		t.Error("expected error when the selected item is gone from the course")
	}
}
