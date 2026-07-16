package botmanager

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"revisitr/internal/entity"
	lunchUC "revisitr/internal/usecase/lunch"
)

// lunchFormatByID returns the format with the given id, or nil.
func lunchFormatByID(program *entity.LunchProgram, formatID int) *entity.LunchFormat {
	for i := range program.Formats {
		if program.Formats[i].ID == formatID {
			return &program.Formats[i]
		}
	}
	return nil
}

// lunchFormatCourses returns the format's courses in the format's order.
func lunchFormatCourses(program *entity.LunchProgram, format *entity.LunchFormat) []entity.LunchCourse {
	byID := make(map[int]entity.LunchCourse, len(program.Courses))
	for _, course := range program.Courses {
		byID[course.ID] = course
	}
	courses := make([]entity.LunchCourse, 0, len(format.CourseIDs))
	for _, courseID := range format.CourseIDs {
		if course, ok := byID[courseID]; ok {
			courses = append(courses, course)
		}
	}
	return courses
}

// lunchFormatPriceLabel renders "350 ₽" for fixed formats and "от N ₽" for
// item-dependent ones.
func lunchFormatPriceLabel(program *entity.LunchProgram, format entity.LunchFormat) string {
	if format.PriceMode == entity.LunchPriceFixed {
		return formatMenuPrice(format.BasePrice)
	}
	min := lunchUC.MinTotal(format, lunchFormatCourses(program, &format))
	return "от " + formatMenuPrice(min)
}

// lunchFormatsContent builds the format selection screen.
func lunchFormatsContent(program *entity.LunchProgram) entity.MessageContent {
	var sb strings.Builder
	sb.WriteString("<b>" + html.EscapeString(program.Name) + "</b>")
	if desc := strings.TrimSpace(program.Description); desc != "" {
		sb.WriteString("\n" + html.EscapeString(desc))
	}
	sb.WriteString("\n\nВыберите формат:")

	var buttons [][]entity.InlineButton
	for _, format := range program.Formats {
		buttons = append(buttons, []entity.InlineButton{{
			Text: format.Name + " — " + lunchFormatPriceLabel(program, format),
			Data: callbackLunchFormatPref + strconv.Itoa(format.ID),
		}})
	}
	buttons = append(buttons, []entity.InlineButton{{Text: "✕ Отмена", Data: callbackLunchCancel}})

	return entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      truncateMenuText(sb.String()),
			ParseMode: "HTML",
		}},
		Buttons: buttons,
	}
}

// lunchSummaryLine renders the live assembly summary, e.g.
// "Собрано: Первое — Борщ · Второе — …".
func lunchSummaryLine(courses []entity.LunchCourse, selections map[int]int) string {
	var parts []string
	for _, course := range courses {
		name := "…"
		if itemID, ok := selections[course.ID]; ok {
			for _, item := range course.Items {
				if item.MenuItemID == itemID && item.MenuItem != nil {
					name = item.MenuItem.Name
					break
				}
			}
		}
		parts = append(parts, course.Title+" — "+name)
	}
	if len(parts) == 0 {
		return ""
	}
	return "Собрано: " + strings.Join(parts, " · ")
}

// lunchCourseContent builds the assembly screen for the current course:
// progress header, item card, nav arrows, select button and course navigation.
func lunchCourseContent(courses []entity.LunchCourse, state FlowState) entity.MessageContent {
	course := courses[state.LunchCourseIdx]
	items := course.Items
	idx := state.LunchItemIdx
	if idx < 0 || idx >= len(items) {
		idx = 0
	}
	item := items[idx]
	selectedItemID := state.LunchSelections[course.ID]

	heading := fmt.Sprintf("Шаг %d/%d · %s", state.LunchCourseIdx+1, len(courses), course.Title)
	text := formatMenuItemCardText(heading, *item.MenuItem)
	if item.Surcharge > 0 {
		text += "\n\nДоплата: +" + formatMenuPrice(item.Surcharge)
	}
	if summary := lunchSummaryLine(courses, state.LunchSelections); summary != "" {
		text += "\n\n<i>" + html.EscapeString(summary) + "</i>"
	}

	var nav []entity.InlineButton
	if idx > 0 {
		nav = append(nav, entity.InlineButton{Text: "←", Data: callbackLunchNavPref + strconv.Itoa(idx - 1)})
	}
	nav = append(nav, entity.InlineButton{
		Text: fmt.Sprintf("%d/%d", idx+1, len(items)),
		Data: callbackLunchNoop,
	})
	if idx < len(items)-1 {
		nav = append(nav, entity.InlineButton{Text: "→", Data: callbackLunchNavPref + strconv.Itoa(idx + 1)})
	}

	selectBtn := entity.InlineButton{
		Text: "✓ Выбрать",
		Data: callbackLunchSelectPref + strconv.Itoa(item.MenuItemID),
	}
	if selectedItemID == item.MenuItemID {
		selectBtn.Text = "✓ Выбрано"
		selectBtn.Style = "success"
	}

	stepNav := []entity.InlineButton{{Text: "← Назад", Data: callbackLunchBack}}
	if selectedItemID != 0 {
		stepNav = append(stepNav, entity.InlineButton{Text: "Далее →", Data: callbackLunchNext})
	}

	buttons := [][]entity.InlineButton{
		nav,
		{selectBtn},
		stepNav,
		{{Text: "✕ Отмена", Data: callbackLunchCancel}},
	}

	if item.MenuItem.ImageURL != nil && *item.MenuItem.ImageURL != "" {
		return entity.MessageContent{
			Parts: []entity.MessagePart{{
				Type:      entity.PartPhoto,
				MediaURL:  *item.MenuItem.ImageURL,
				Text:      truncateMenuCaption(text),
				ParseMode: "HTML",
			}},
			Buttons: buttons,
		}
	}
	return entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      truncateMenuText(text),
			ParseMode: "HTML",
		}},
		Buttons: buttons,
	}
}

// lunchConfirmContent builds the confirmation screen with per-line edit
// buttons, the table number and the total.
func lunchConfirmContent(format *entity.LunchFormat, courses []entity.LunchCourse, state FlowState, total float64) entity.MessageContent {
	var sb strings.Builder
	sb.WriteString("<b>Ваш заказ · " + html.EscapeString(format.Name) + "</b>\n\n")

	var buttons [][]entity.InlineButton
	for i, course := range courses {
		name := "—"
		price := ""
		if itemID, ok := state.LunchSelections[course.ID]; ok {
			for _, item := range course.Items {
				if item.MenuItemID == itemID && item.MenuItem != nil {
					name = item.MenuItem.Name
					switch format.PriceMode {
					case entity.LunchPriceSumOfItems:
						price = " · " + formatMenuPrice(item.MenuItem.Price)
					case entity.LunchPriceBasePlusSurcharge:
						if item.Surcharge > 0 {
							price = " · +" + formatMenuPrice(item.Surcharge)
						}
					}
					break
				}
			}
		}
		sb.WriteString(html.EscapeString(course.Title) + ": " + html.EscapeString(name) + html.EscapeString(price) + "\n")
		buttons = append(buttons, []entity.InlineButton{{
			Text: "Изменить: " + course.Title,
			Data: callbackLunchEditPref + strconv.Itoa(i),
		}})
	}

	sb.WriteString("\nСтол: <b>" + html.EscapeString(state.TableNum) + "</b>")
	sb.WriteString("\nИтого: <b>" + formatMenuPrice(total) + "</b>")

	buttons = append(buttons,
		[]entity.InlineButton{{Text: "Стол: " + state.TableNum + " · изменить", Data: callbackLunchTable}},
		[]entity.InlineButton{
			{Text: "✅ Оформить", Data: callbackLunchConfirm, Style: "success"},
			{Text: "✕ Отмена", Data: callbackLunchCancel},
		},
	)

	return entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      truncateMenuText(sb.String()),
			ParseMode: "HTML",
		}},
		Buttons: buttons,
	}
}

// lunchPriceInput assembles the price calculation input from the guest's
// selections, in course order.
func lunchPriceInput(format *entity.LunchFormat, courses []entity.LunchCourse, selections map[int]int) lunchUC.LunchPriceInput {
	in := lunchUC.LunchPriceInput{
		PriceMode: format.PriceMode,
		BasePrice: format.BasePrice,
	}
	for _, course := range courses {
		itemID, ok := selections[course.ID]
		if !ok {
			continue
		}
		for _, item := range course.Items {
			if item.MenuItemID == itemID && item.MenuItem != nil {
				in.Items = append(in.Items, lunchUC.LunchPriceItem{
					MenuItemPrice: item.MenuItem.Price,
					Surcharge:     item.Surcharge,
				})
				break
			}
		}
	}
	return in
}
