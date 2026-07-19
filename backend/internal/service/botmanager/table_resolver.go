package botmanager

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
)

// Table resolution is a generic flow step shared across modules. The only
// method for now is manual typing; future methods (QR deep-link ?table=NN,
// spatial detection) pre-fill state.TableNum and skip the prompt entirely.

// startTableResolve asks the guest for their table number. The requesting
// flow is recorded in AwaitingTableFor so handleTableInput can route back.
func (h *handler) startTableResolve(ctx context.Context, chatID int64, state FlowState, forFlow string) {
	state.AwaitingTableFor = forFlow
	state.FlowMessageID = h.replaceFlowMessage(ctx, chatID, state, entity.MessageContent{
		Parts: []entity.MessagePart{{
			Type: entity.PartText,
			Text: "За каким столом вы сидите? Отправьте номер стола сообщением.",
		}},
		Buttons: [][]entity.InlineButton{
			{{Text: "✕ Отмена", Data: callbackLunchCancel}},
		},
	})
	_ = h.saveFlowState(ctx, chatID, state)
}

// handleTableInput consumes a typed table number and returns the guest to
// the flow that requested it.
func (h *handler) handleTableInput(ctx context.Context, msg *telego.Message, text string, state FlowState) {
	chatID := msg.Chat.ID

	tableNum, err := validateTableNum(text)
	if err != nil {
		h.sendText(chatID, err.Error())
		return
	}

	forFlow := state.AwaitingTableFor
	state.AwaitingTableFor = ""
	state.TableNum = tableNum
	_ = h.saveFlowState(ctx, chatID, state)

	switch forFlow {
	case "lunch":
		program := h.lunchData(ctx)
		if program == nil {
			h.sendText(chatID, "Ланч сейчас недоступен.")
			return
		}
		h.renderLunchConfirm(ctx, chatID, program, state)
	default:
		h.logger.Warn("table input for unknown flow", "flow", forFlow)
	}
}

func validateTableNum(text string) (string, error) {
	table := strings.TrimSpace(text)
	runes := []rune(table)
	if len(runes) == 0 || len(runes) > 8 {
		return "", errors.New("Номер стола — до 8 букв или цифр. Попробуйте ещё раз.")
	}
	for _, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", errors.New("Номер стола — только буквы и цифры. Попробуйте ещё раз.")
		}
	}
	return table, nil
}
