package botmanager

import (
	"strings"
	"testing"

	"revisitr/internal/entity"
)

func lunchTestProgram() *entity.LunchProgram {
	img := "https://example.com/borsch.jpg"
	return &entity.LunchProgram{
		ID:       1,
		Name:     "Бизнес-ланч",
		IsActive: true,
		Courses: []entity.LunchCourse{
			{
				ID: 10, Code: "1", Title: "Первое",
				Items: []entity.LunchCourseItem{
					{CourseID: 10, MenuItemID: 100, MenuItem: &entity.MenuItem{ID: 100, Name: "Борщ", Price: 180, IsAvailable: true, ImageURL: &img}},
					{CourseID: 10, MenuItemID: 101, MenuItem: &entity.MenuItem{ID: 101, Name: "Солянка", Price: 220, IsAvailable: true}},
				},
			},
			{
				ID: 20, Code: "2", Title: "Второе",
				Items: []entity.LunchCourseItem{
					{CourseID: 20, MenuItemID: 200, Surcharge: 200, MenuItem: &entity.MenuItem{ID: 200, Name: "Стейк", Price: 550, IsAvailable: true}},
					{CourseID: 20, MenuItemID: 201, MenuItem: &entity.MenuItem{ID: 201, Name: "Котлета", Price: 250, IsAvailable: true}},
				},
			},
		},
		Formats: []entity.LunchFormat{
			{ID: 5, Name: "Первое + Второе", PriceMode: entity.LunchPriceFixed, BasePrice: 350, IsActive: true, CourseIDs: []int{10, 20}},
			{ID: 6, Name: "Только второе", PriceMode: entity.LunchPriceSumOfItems, IsActive: true, CourseIDs: []int{20}},
		},
	}
}

func allCallbackData(content entity.MessageContent) []string {
	var data []string
	for _, row := range content.Buttons {
		for _, btn := range row {
			if btn.Data != "" {
				data = append(data, btn.Data)
			}
		}
	}
	return data
}

func findButton(content entity.MessageContent, text string) *entity.InlineButton {
	for _, row := range content.Buttons {
		for i := range row {
			if strings.HasPrefix(row[i].Text, text) {
				return &row[i]
			}
		}
	}
	return nil
}

func TestLunchFormatsContentPriceLabels(t *testing.T) {
	program := lunchTestProgram()
	content := lunchFormatsContent(program)

	fixed := findButton(content, "Первое + Второе")
	if fixed == nil {
		t.Fatal("fixed format button missing")
	}
	if !strings.Contains(fixed.Text, "350") || strings.Contains(fixed.Text, "от") {
		t.Errorf("fixed format must show exact price, got %q", fixed.Text)
	}

	byItems := findButton(content, "Только второе")
	if byItems == nil {
		t.Fatal("by-items format button missing")
	}
	if !strings.Contains(byItems.Text, "от") || !strings.Contains(byItems.Text, "250") {
		t.Errorf("by-items format must show 'от <min>' (cheapest = 250), got %q", byItems.Text)
	}
}

func lunchTestState() FlowState {
	return FlowState{
		Flow:          "lunch",
		LunchFormatID: 5,
	}
}

func TestLunchCourseContentProgressHeader(t *testing.T) {
	program := lunchTestProgram()
	state := lunchTestState()
	state.LunchCourseIdx = 1
	courses := lunchFormatCourses(program, lunchFormatByID(program, 5))

	content := lunchCourseContent(courses, state)
	text := content.Parts[0].Text
	if !strings.Contains(text, "Шаг 2/2 · Второе") {
		t.Errorf("progress header missing, got %q", text)
	}
}

func TestLunchCourseContentSelectHighlight(t *testing.T) {
	program := lunchTestProgram()
	courses := lunchFormatCourses(program, lunchFormatByID(program, 5))

	state := lunchTestState()
	content := lunchCourseContent(courses, state)
	btn := findButton(content, "✓ Выбрать")
	if btn == nil || btn.Style == "success" {
		t.Errorf("unselected item must show plain select button, got %+v", btn)
	}

	state.LunchSelections = map[int]int{10: 100}
	content = lunchCourseContent(courses, state)
	btn = findButton(content, "✓ Выбрано")
	if btn == nil || btn.Style != "success" {
		t.Errorf("selected item must show green button, got %+v", btn)
	}
}

func TestLunchCourseContentNextGating(t *testing.T) {
	program := lunchTestProgram()
	courses := lunchFormatCourses(program, lunchFormatByID(program, 5))

	state := lunchTestState()
	content := lunchCourseContent(courses, state)
	if findButton(content, "Далее") != nil {
		t.Error("Далее must be absent before a selection is made")
	}

	state.LunchSelections = map[int]int{10: 101}
	content = lunchCourseContent(courses, state)
	if findButton(content, "Далее") == nil {
		t.Error("Далее must appear once the course has a selection")
	}
}

func TestLunchCourseContentNavBounds(t *testing.T) {
	program := lunchTestProgram()
	courses := lunchFormatCourses(program, lunchFormatByID(program, 5))
	state := lunchTestState()

	// First item: no ← arrow.
	content := lunchCourseContent(courses, state)
	for _, data := range allCallbackData(content) {
		if data == callbackLunchNavPref+"-1" {
			t.Error("left arrow must be absent on the first item")
		}
	}
	if findButton(content, "→") == nil {
		t.Error("right arrow must be present on the first item")
	}

	// Last item: no → arrow.
	state.LunchItemIdx = 1
	content = lunchCourseContent(courses, state)
	for _, data := range allCallbackData(content) {
		if data == callbackLunchNavPref+"2" {
			t.Error("right arrow must be absent on the last item")
		}
	}
	if findButton(content, "←") == nil {
		t.Error("left arrow must be present on the last item")
	}
}

func TestLunchCourseContentPhotoVsText(t *testing.T) {
	program := lunchTestProgram()
	courses := lunchFormatCourses(program, lunchFormatByID(program, 5))
	state := lunchTestState()

	content := lunchCourseContent(courses, state)
	if content.Parts[0].Type != entity.PartPhoto {
		t.Errorf("item with image must render as photo, got %q", content.Parts[0].Type)
	}

	state.LunchItemIdx = 1 // Солянка has no image
	content = lunchCourseContent(courses, state)
	if content.Parts[0].Type != entity.PartText {
		t.Errorf("item without image must render as text, got %q", content.Parts[0].Type)
	}
}

func TestLunchSummaryLine(t *testing.T) {
	program := lunchTestProgram()
	courses := lunchFormatCourses(program, lunchFormatByID(program, 5))

	got := lunchSummaryLine(courses, map[int]int{10: 100})
	want := "Собрано: Первое — Борщ · Второе — …"
	if got != want {
		t.Errorf("summary: got %q, want %q", got, want)
	}
}

func TestLunchConfirmContent(t *testing.T) {
	program := lunchTestProgram()
	format := lunchFormatByID(program, 5)
	courses := lunchFormatCourses(program, format)
	state := lunchTestState()
	state.LunchSelections = map[int]int{10: 100, 20: 201}
	state.TableNum = "7"

	content := lunchConfirmContent(format, courses, state, 350)
	text := content.Parts[0].Text

	for _, want := range []string{"Борщ", "Котлета", "Стол: <b>7</b>", "Итого"} {
		if !strings.Contains(text, want) {
			t.Errorf("confirm text missing %q, got %q", want, text)
		}
	}

	if findButton(content, "Изменить: Первое") == nil || findButton(content, "Изменить: Второе") == nil {
		t.Error("per-course edit buttons missing")
	}
	if findButton(content, "✅ Оформить") == nil {
		t.Error("confirm button missing")
	}
}

func TestLunchPriceInputAssembly(t *testing.T) {
	program := lunchTestProgram()
	format := lunchFormatByID(program, 5)
	courses := lunchFormatCourses(program, format)

	in := lunchPriceInput(format, courses, map[int]int{10: 101, 20: 200})
	if len(in.Items) != 2 {
		t.Fatalf("expected 2 price items, got %d", len(in.Items))
	}
	if in.Items[0].MenuItemPrice != 220 || in.Items[1].Surcharge != 200 {
		t.Errorf("price input mismatch: %+v", in.Items)
	}
}

func TestLunchCallbackDataWithin64Bytes(t *testing.T) {
	program := lunchTestProgram()
	format := lunchFormatByID(program, 5)
	courses := lunchFormatCourses(program, format)
	state := lunchTestState()
	state.LunchSelections = map[int]int{10: 100, 20: 201}
	state.TableNum = "12345678"

	contents := []entity.MessageContent{
		lunchFormatsContent(program),
		lunchCourseContent(courses, state),
		lunchConfirmContent(format, courses, state, 350),
	}
	for _, content := range contents {
		for _, data := range allCallbackData(content) {
			if len(data) > 64 {
				t.Errorf("callback data %q exceeds Telegram's 64-byte limit (%d)", data, len(data))
			}
		}
	}
}
