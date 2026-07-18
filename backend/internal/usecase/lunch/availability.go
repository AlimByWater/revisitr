package lunch

import (
	"fmt"
	"strings"
	"time"

	"revisitr/internal/entity"
)

// isoWeekday converts Go's time.Weekday (Sunday = 0) to ISO (Monday = 1 … Sunday = 7).
func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

// IsAvailableAt reports whether t falls inside any availability slot.
// Slot start is inclusive, end is exclusive (12:00–16:00 admits 12:00, not 16:00).
// Times are evaluated in t's location; callers pass the current time in the
// organization's timezone (see handler.orgNow in botmanager).
func IsAvailableAt(slots []entity.LunchAvailability, t time.Time) bool {
	weekday := isoWeekday(t)
	minutes := t.Hour()*60 + t.Minute()
	for _, slot := range slots {
		if slot.Weekday != weekday {
			continue
		}
		from, errFrom := parseMinutes(slot.TimeFrom)
		to, errTo := parseMinutes(slot.TimeTo)
		if errFrom != nil || errTo != nil {
			continue
		}
		if minutes >= from && minutes < to {
			return true
		}
	}
	return false
}

func parseMinutes(s string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d:%d", &h, &m); err != nil {
		return 0, err
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time %q", s)
	}
	return h*60 + m, nil
}

var weekdayShortNames = [8]string{"", "пн", "вт", "ср", "чт", "пт", "сб", "вс"}

// FormatSchedule renders slots as a human-readable schedule for the bot's
// polite refusal message, e.g. "пн–пт 12:00–16:00, сб 12:00–15:00".
func FormatSchedule(slots []entity.LunchAvailability) string {
	if len(slots) == 0 {
		return ""
	}

	// Group consecutive weekdays sharing the same time range, preserving
	// weekday order.
	type timeRange struct{ from, to string }
	byDay := map[int]timeRange{}
	for _, slot := range slots {
		if slot.Weekday < 1 || slot.Weekday > 7 {
			continue
		}
		// One slot per weekday is the admin UI contract; last one wins otherwise.
		byDay[slot.Weekday] = timeRange{from: slot.TimeFrom, to: slot.TimeTo}
	}

	var parts []string
	for day := 1; day <= 7; day++ {
		tr, ok := byDay[day]
		if !ok {
			continue
		}
		end := day
		for end+1 <= 7 {
			next, nextOK := byDay[end+1]
			if !nextOK || next != tr {
				break
			}
			end++
		}
		label := weekdayShortNames[day]
		if end > day {
			label += "–" + weekdayShortNames[end]
		}
		parts = append(parts, fmt.Sprintf("%s %s–%s", label, tr.from, tr.to))
		day = end
	}
	return strings.Join(parts, ", ")
}
