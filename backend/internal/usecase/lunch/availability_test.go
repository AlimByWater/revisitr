package lunch

import (
	"testing"
	"time"

	"revisitr/internal/entity"
)

// 2026-07-13 is a Monday.
func mondayAt(hour, minute int) time.Time {
	return time.Date(2026, 7, 13, hour, minute, 0, 0, time.UTC)
}

func weekdaySlots() []entity.LunchAvailability {
	return []entity.LunchAvailability{
		{Weekday: 1, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 2, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 7, TimeFrom: "10:00", TimeTo: "14:00"},
	}
}

func TestIsAvailableAtInsideWindow(t *testing.T) {
	if !IsAvailableAt(weekdaySlots(), mondayAt(13, 30)) {
		t.Error("monday 13:30 must be available")
	}
}

func TestIsAvailableAtOutsideWindow(t *testing.T) {
	if IsAvailableAt(weekdaySlots(), mondayAt(9, 0)) {
		t.Error("monday 09:00 must not be available")
	}
	if IsAvailableAt(weekdaySlots(), mondayAt(18, 0)) {
		t.Error("monday 18:00 must not be available")
	}
}

func TestIsAvailableAtBoundaries(t *testing.T) {
	if !IsAvailableAt(weekdaySlots(), mondayAt(12, 0)) {
		t.Error("window start 12:00 must be inclusive")
	}
	if IsAvailableAt(weekdaySlots(), mondayAt(16, 0)) {
		t.Error("window end 16:00 must be exclusive")
	}
	if !IsAvailableAt(weekdaySlots(), mondayAt(15, 59)) {
		t.Error("15:59 must be available")
	}
}

func TestIsAvailableAtWeekdayMismatch(t *testing.T) {
	wednesday := time.Date(2026, 7, 15, 13, 0, 0, 0, time.UTC)
	if IsAvailableAt(weekdaySlots(), wednesday) {
		t.Error("wednesday must not be available (no slot)")
	}
}

func TestIsAvailableAtSundayISOConversion(t *testing.T) {
	// 2026-07-19 is a Sunday; Go's Weekday() returns 0, ISO slot uses 7.
	sunday := time.Date(2026, 7, 19, 11, 0, 0, 0, time.UTC)
	if !IsAvailableAt(weekdaySlots(), sunday) {
		t.Error("sunday 11:00 must match the weekday=7 slot")
	}
}

func TestIsAvailableAtEmptySlots(t *testing.T) {
	if IsAvailableAt(nil, mondayAt(13, 0)) {
		t.Error("no slots means never available")
	}
}

func TestFormatScheduleGroupsConsecutiveDays(t *testing.T) {
	slots := []entity.LunchAvailability{
		{Weekday: 1, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 2, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 3, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 4, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 5, TimeFrom: "12:00", TimeTo: "16:00"},
		{Weekday: 6, TimeFrom: "12:00", TimeTo: "15:00"},
	}
	got := FormatSchedule(slots)
	want := "пн–пт 12:00–16:00, сб 12:00–15:00"
	if got != want {
		t.Errorf("FormatSchedule: got %q, want %q", got, want)
	}
}

func TestFormatScheduleEmpty(t *testing.T) {
	if got := FormatSchedule(nil); got != "" {
		t.Errorf("empty slots: got %q, want empty string", got)
	}
}
