package botmanager

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"revisitr/internal/entity"
	"revisitr/internal/service/eventbus"
	lunchUC "revisitr/internal/usecase/lunch"

	"github.com/mymmrac/telego"
)

const (
	callbackLunchPrefix     = "lunch:"
	callbackLunchStart      = "lunch:start"
	callbackLunchFormatPref = "lunch:fmt:"
	callbackLunchNavPref    = "lunch:nav:"
	callbackLunchSelectPref = "lunch:sel:"
	callbackLunchNext       = "lunch:next"
	callbackLunchBack       = "lunch:back"
	callbackLunchEditPref   = "lunch:edit:"
	callbackLunchTable      = "lunch:table"
	callbackLunchConfirm    = "lunch:ok"
	callbackLunchCancel     = "lunch:cancel"
	callbackLunchNoop       = "lunch:noop"
)

func (h *handler) lunchModuleEnabled() bool {
	for _, module := range h.info.Settings.Modules {
		if module == "lunch" {
			return true
		}
	}
	return false
}

// lunchData loads the bot's lunch program and applies FR-B6 filtering:
// unavailable items are dropped, emptied courses are dropped, and inactive
// formats or formats referencing a dropped course are dropped.
func (h *handler) lunchData(ctx context.Context) *entity.LunchProgram {
	if h.mgr.lunchRepo == nil {
		return nil
	}
	program, err := h.mgr.lunchRepo.GetFullProgramByBotID(ctx, h.info.ID)
	if err != nil {
		h.logger.Error("load lunch program", "error", err)
		return nil
	}
	if program == nil || !program.IsActive {
		return nil
	}

	courseIDs := map[int]bool{}
	courses := make([]entity.LunchCourse, 0, len(program.Courses))
	for _, course := range program.Courses {
		var items []entity.LunchCourseItem
		for _, item := range course.Items {
			if item.MenuItem == nil || !item.MenuItem.IsAvailable {
				continue
			}
			items = append(items, item)
		}
		if len(items) == 0 {
			continue
		}
		course.Items = items
		courses = append(courses, course)
		courseIDs[course.ID] = true
	}
	program.Courses = courses

	formats := make([]entity.LunchFormat, 0, len(program.Formats))
	for _, format := range program.Formats {
		if !format.IsActive || len(format.CourseIDs) == 0 {
			continue
		}
		complete := true
		for _, courseID := range format.CourseIDs {
			if !courseIDs[courseID] {
				complete = false
				break
			}
		}
		if complete {
			formats = append(formats, format)
		}
	}
	program.Formats = formats

	return program
}

// orgNow returns the current time in the organization's timezone, falling
// back to server time when the timezone is unknown or unresolvable.
func (h *handler) orgNow(ctx context.Context) time.Time {
	now := time.Now()
	if h.mgr.orgsRepo == nil {
		return now
	}
	tz, err := h.mgr.orgsRepo.GetTimezone(ctx, h.info.OrgID)
	if err != nil || tz == "" {
		if err != nil {
			h.logger.Warn("org timezone lookup failed", "error", err, "org_id", h.info.OrgID)
		}
		return now
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		h.logger.Warn("org timezone invalid", "timezone", tz, "org_id", h.info.OrgID)
		return now
	}
	return now.In(loc)
}

// lunchAvailableNow reports whether the lunch module should be visible:
// program active, has formats and the current org-local time is inside
// an availability window.
func (h *handler) lunchAvailableNow(ctx context.Context) bool {
	program := h.lunchData(ctx)
	return program != nil && len(program.Formats) > 0 &&
		lunchUC.IsAvailableAt(program.Availability, h.orgNow(ctx))
}

// mainMenuKeyboard builds the main reply keyboard, hiding the lunch button
// outside its availability window. Reply keyboards go stale by design, so
// handleLunch re-checks the window on entry.
func (h *handler) mainMenuKeyboard(ctx context.Context) *telego.ReplyKeyboardMarkup {
	var hidden map[string]bool
	if h.lunchModuleEnabled() && !h.lunchAvailableNow(ctx) {
		hidden = map[string]bool{"lunch": true}
	}
	return buildMainMenuFiltered(h.info.Settings, hidden)
}

func (h *handler) resetLunchState(state *FlowState) {
	state.LunchFormatID = 0
	state.LunchCourseIdx = 0
	state.LunchItemIdx = 0
	state.LunchSelections = nil
	state.AwaitingTableFor = ""
	state.TableNum = ""
}

// handleLunch is the entry point from the "Ланч" reply button.
func (h *handler) handleLunch(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID

	if !h.lunchModuleEnabled() {
		h.sendText(chatID, "Ланч сейчас недоступен.")
		return
	}
	program := h.lunchData(ctx)
	if program == nil || len(program.Formats) == 0 {
		h.sendText(chatID, "Ланч сейчас недоступен.")
		return
	}
	if !lunchUC.IsAvailableAt(program.Availability, h.orgNow(ctx)) {
		if schedule := lunchUC.FormatSchedule(program.Availability); schedule != "" {
			h.sendText(chatID, "Ланч доступен: "+schedule+". Ждём вас!")
		} else {
			h.sendText(chatID, "Ланч сейчас недоступен.")
		}
		return
	}

	state := h.currentFlowState(ctx, chatID)
	h.resetLunchState(&state)
	state.Flow = "lunch"
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, lunchFormatsContent(program))
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) handleLunchCallback(ctx context.Context, query *telego.CallbackQuery) {
	chatID := query.Message.GetChat().ID
	data := query.Data

	if data == callbackLunchNoop {
		h.answerCallback(query.ID, "")
		return
	}
	if data == callbackLunchCancel {
		h.answerCallback(query.ID, "")
		h.cancelLunchFlow(ctx, chatID)
		return
	}
	program := h.lunchData(ctx)
	if program == nil || len(program.Formats) == 0 {
		h.answerCallback(query.ID, "Ланч сейчас недоступен")
		return
	}
	state := h.currentFlowState(ctx, chatID)

	switch {
	case data == callbackLunchStart:
		h.answerCallback(query.ID, "")
		h.resetLunchState(&state)
		state.Flow = "lunch"
		state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, lunchFormatsContent(program))
		_ = h.saveFlowState(ctx, chatID, state)

	case strings.HasPrefix(data, callbackLunchFormatPref):
		h.answerCallback(query.ID, "")
		formatID, err := strconv.Atoi(strings.TrimPrefix(data, callbackLunchFormatPref))
		if err != nil || lunchFormatByID(program, formatID) == nil {
			h.answerCallback(query.ID, "Действие устарело")
			return
		}
		h.resetLunchState(&state)
		state.Flow = "lunch"
		state.LunchFormatID = formatID
		h.renderLunchCourse(ctx, chatID, program, state, true)

	case strings.HasPrefix(data, callbackLunchNavPref):
		h.answerCallback(query.ID, "")
		idx, err := strconv.Atoi(strings.TrimPrefix(data, callbackLunchNavPref))
		if err != nil {
			return
		}
		state.LunchItemIdx = idx
		// The card photo may change — replace instead of edit.
		h.renderLunchCourse(ctx, chatID, program, state, true)

	case strings.HasPrefix(data, callbackLunchSelectPref):
		h.answerCallback(query.ID, "")
		itemID, err := strconv.Atoi(strings.TrimPrefix(data, callbackLunchSelectPref))
		if err != nil {
			return
		}
		courses, ok := h.lunchCurrentCourses(program, &state)
		if !ok {
			return
		}
		if state.LunchSelections == nil {
			state.LunchSelections = map[int]int{}
		}
		state.LunchSelections[courses[state.LunchCourseIdx].ID] = itemID
		// Same card, same photo — edit in place for a flicker-free toggle.
		h.renderLunchCourse(ctx, chatID, program, state, false)

	case data == callbackLunchNext:
		courses, ok := h.lunchCurrentCourses(program, &state)
		if !ok {
			h.answerCallback(query.ID, "Действие устарело")
			return
		}
		if state.LunchSelections[courses[state.LunchCourseIdx].ID] == 0 {
			h.answerCallback(query.ID, "Сначала выберите позицию")
			return
		}
		h.answerCallback(query.ID, "")
		if state.LunchCourseIdx == len(courses)-1 {
			if state.TableNum == "" {
				h.startTableResolve(ctx, chatID, state, "lunch")
				return
			}
			h.renderLunchConfirm(ctx, chatID, program, state)
			return
		}
		state.LunchCourseIdx++
		state.LunchItemIdx = lunchSelectedItemIdx(courses[state.LunchCourseIdx], state.LunchSelections)
		h.renderLunchCourse(ctx, chatID, program, state, true)

	case data == callbackLunchBack:
		h.answerCallback(query.ID, "")
		courses, ok := h.lunchCurrentCourses(program, &state)
		if !ok {
			return
		}
		if state.LunchCourseIdx == 0 {
			state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, lunchFormatsContent(program))
			_ = h.saveFlowState(ctx, chatID, state)
			return
		}
		state.LunchCourseIdx--
		state.LunchItemIdx = lunchSelectedItemIdx(courses[state.LunchCourseIdx], state.LunchSelections)
		h.renderLunchCourse(ctx, chatID, program, state, true)

	case strings.HasPrefix(data, callbackLunchEditPref):
		h.answerCallback(query.ID, "")
		idx, err := strconv.Atoi(strings.TrimPrefix(data, callbackLunchEditPref))
		if err != nil {
			return
		}
		courses, ok := h.lunchCurrentCourses(program, &state)
		if !ok || idx < 0 || idx >= len(courses) {
			return
		}
		state.LunchCourseIdx = idx
		state.LunchItemIdx = lunchSelectedItemIdx(courses[idx], state.LunchSelections)
		h.renderLunchCourse(ctx, chatID, program, state, true)

	case data == callbackLunchTable:
		h.answerCallback(query.ID, "")
		state.TableNum = ""
		h.startTableResolve(ctx, chatID, state, "lunch")

	case data == callbackLunchConfirm:
		h.finalizeLunchOrder(ctx, query, program, state)

	default:
		h.answerCallback(query.ID, "Действие устарело")
	}
}

// lunchCurrentCourses resolves the state's format into its ordered course
// list, clamping the course index against config changes mid-flow.
func (h *handler) lunchCurrentCourses(program *entity.LunchProgram, state *FlowState) ([]entity.LunchCourse, bool) {
	format := lunchFormatByID(program, state.LunchFormatID)
	if format == nil {
		return nil, false
	}
	courses := lunchFormatCourses(program, format)
	if len(courses) == 0 {
		return nil, false
	}
	if state.LunchCourseIdx < 0 || state.LunchCourseIdx >= len(courses) {
		state.LunchCourseIdx = 0
	}
	return courses, true
}

// lunchSelectedItemIdx returns the carousel index of the course's selected
// item, or 0 when nothing is selected yet.
func lunchSelectedItemIdx(course entity.LunchCourse, selections map[int]int) int {
	itemID, ok := selections[course.ID]
	if !ok {
		return 0
	}
	for i, item := range course.Items {
		if item.MenuItemID == itemID {
			return i
		}
	}
	return 0
}

func (h *handler) renderLunchCourse(ctx context.Context, chatID int64, program *entity.LunchProgram, state FlowState, replace bool) {
	courses, ok := h.lunchCurrentCourses(program, &state)
	if !ok {
		state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, lunchFormatsContent(program))
		_ = h.saveFlowState(ctx, chatID, state)
		return
	}
	if state.LunchItemIdx < 0 || state.LunchItemIdx >= len(courses[state.LunchCourseIdx].Items) {
		state.LunchItemIdx = 0
	}

	content := lunchCourseContent(courses, state)
	if replace {
		state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, content)
	} else {
		state.FlowMessageID = h.updateFlowMessage(ctx, chatID, state, content)
	}
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) renderLunchConfirm(ctx context.Context, chatID int64, program *entity.LunchProgram, state FlowState) {
	format := lunchFormatByID(program, state.LunchFormatID)
	if format == nil {
		state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, lunchFormatsContent(program))
		_ = h.saveFlowState(ctx, chatID, state)
		return
	}
	courses := lunchFormatCourses(program, format)

	total, err := lunchUC.CalculateTotal(lunchPriceInput(format, courses, state.LunchSelections))
	if err != nil {
		h.logger.Error("lunch total", "error", err, "format_id", format.ID)
		total = 0
	}

	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, lunchConfirmContent(format, courses, state, total))
	_ = h.saveFlowState(ctx, chatID, state)
}

// buildLunchOrder assembles an order with text snapshots and the calculated
// total from the guest's selections. Pure — unit-testable.
func buildLunchOrder(botID, clientID int, format *entity.LunchFormat, courses []entity.LunchCourse, selections map[int]int, tableNum string) (*entity.Order, error) {
	total, err := lunchUC.CalculateTotal(lunchPriceInput(format, courses, selections))
	if err != nil {
		return nil, err
	}

	formatID := format.ID
	order := &entity.Order{
		BotID:       botID,
		BotClientID: clientID,
		Source:      entity.OrderSourceLunch,
		FormatID:    &formatID,
		FormatName:  format.Name,
		TableNum:    tableNum,
		TotalPrice:  total,
		Status:      entity.OrderStatusNew,
	}

	for _, course := range courses {
		itemID, ok := selections[course.ID]
		if !ok {
			return nil, fmt.Errorf("course %q has no selection", course.Title)
		}
		var selected *entity.LunchCourseItem
		for i := range course.Items {
			if course.Items[i].MenuItemID == itemID {
				selected = &course.Items[i]
				break
			}
		}
		if selected == nil || selected.MenuItem == nil {
			return nil, fmt.Errorf("selected item %d is gone from course %q", itemID, course.Title)
		}
		courseID := course.ID
		menuItemID := selected.MenuItemID
		order.Items = append(order.Items, entity.OrderLine{
			CourseID:    &courseID,
			CourseTitle: course.Title,
			MenuItemID:  &menuItemID,
			ItemName:    selected.MenuItem.Name,
			Price:       selected.MenuItem.Price,
			Surcharge:   selected.Surcharge,
		})
	}

	return order, nil
}

// finalizeLunchOrder re-validates the assembly and persists the order with a
// price snapshot. The order lands in the admin panel (status "new"); Telegram
// delivery to staff is a future upgrade pending the waiters-chat decision.
func (h *handler) finalizeLunchOrder(ctx context.Context, query *telego.CallbackQuery, program *entity.LunchProgram, state FlowState) {
	chatID := query.Message.GetChat().ID

	format := lunchFormatByID(program, state.LunchFormatID)
	if format == nil {
		h.answerCallback(query.ID, "Действие устарело")
		return
	}
	courses := lunchFormatCourses(program, format)

	// The window may have closed while the guest was assembling.
	if !lunchUC.IsAvailableAt(program.Availability, h.orgNow(ctx)) {
		h.answerCallback(query.ID, "Ланч уже закрылся")
		if schedule := lunchUC.FormatSchedule(program.Availability); schedule != "" {
			h.sendText(chatID, "Ланч уже закрылся. Расписание: "+schedule+".")
		}
		return
	}

	if state.TableNum == "" {
		h.answerCallback(query.ID, "")
		h.startTableResolve(ctx, chatID, state, "lunch")
		return
	}

	// Stop-list race: a selected item may have become unavailable —
	// lunchData already filtered it out, so buildLunchOrder will catch it.
	if h.mgr.ordersRepo == nil {
		h.answerCallback(query.ID, "Ланч сейчас недоступен")
		return
	}

	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, query.From.ID)
	if err != nil || client == nil {
		h.answerCallback(query.ID, "")
		h.sendText(chatID, "Не нашли вашу регистрацию — отправьте /start и соберите заказ заново.")
		return
	}

	order, err := buildLunchOrder(h.info.ID, client.ID, format, courses, state.LunchSelections, state.TableNum)
	if err != nil {
		h.logger.Warn("lunch order build failed", "error", err, "chat_id", chatID)
		h.answerCallback(query.ID, "Часть позиций стала недоступна — проверьте состав")
		state.LunchCourseIdx = 0
		state.LunchItemIdx = 0
		h.renderLunchCourse(ctx, chatID, program, state, true)
		return
	}

	if err := h.mgr.ordersRepo.Create(ctx, order); err != nil {
		h.logger.Error("lunch order create failed", "error", err, "chat_id", chatID)
		h.answerCallback(query.ID, "Не удалось оформить заказ, попробуйте ещё раз")
		return
	}

	if h.mgr.lunchEvents != nil {
		if err := h.mgr.lunchEvents.PublishLunchOrderCreated(ctx, eventbus.LunchOrderEvent{
			OrderID:  order.ID,
			BotID:    order.BotID,
			TableNum: order.TableNum,
			Total:    order.TotalPrice,
		}); err != nil {
			// Never block the guest on event delivery.
			h.logger.Error("lunch order event publish failed", "error", err, "order_id", order.ID)
		}
	}

	h.answerCallback(query.ID, "")
	h.logger.Info("lunch order created", "order_id", order.ID, "table", order.TableNum, "total", order.TotalPrice)

	confirmation := fmt.Sprintf(
		"✅ <b>Заказ №%d оформлен!</b>\n\nСтол: <b>%s</b>\nСумма: <b>%s</b>\n\nОфициант скоро подойдёт.",
		order.ID, order.TableNum, formatMenuPrice(order.TotalPrice),
	)
	h.resetLunchState(&state)
	state.Flow = ""
	state.FlowMessageID = h.updateFlowMessage(ctx, chatID, state, entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type:      entity.PartText,
			Text:      confirmation,
			ParseMode: "HTML",
		}},
	})
	_ = h.saveFlowState(ctx, chatID, state)
}

func (h *handler) cancelLunchFlow(ctx context.Context, chatID int64) {
	state := h.currentFlowState(ctx, chatID)
	h.resetLunchState(&state)
	state.Flow = ""
	state.FlowMessageID = h.updateFlowMessage(ctx, chatID, state, entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "Сборка ланча отменена.",
		}},
	})
	_ = h.saveFlowState(ctx, chatID, state)
}
