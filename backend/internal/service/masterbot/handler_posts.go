package masterbot

import (
	"context"
	"fmt"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// handleCreatePost processes non-command messages from linked users → creates post codes.
func (h *handler) handleCreatePost(ctx context.Context, msg *telego.Message) {
	if h.deps.PostCodes == nil || h.deps.MasterLinks == nil {
		return
	}

	// Resolve org from master bot link
	masterLink, err := h.deps.MasterLinks.GetLinkByTelegramID(ctx, msg.From.ID)
	if err != nil || masterLink == nil {
		// Not linked via master bot — try legacy
		link := h.getLink(ctx, msg.From.ID)
		if link == nil {
			return // not linked at all, ignore
		}
		// Legacy user — show hint about post creation
		h.sendText(msg.Chat.ID, "💡 Для создания поста привяжите аккаунт через сайт (кнопка «Активировать бота»).")
		return
	}

	// Build post content from message
	content := h.extractContent(msg)
	if content.Text == "" && len(content.MediaURLs) == 0 {
		return // empty message, ignore
	}

	// Generate unique code
	code := entity.GeneratePostCode()

	pc := &entity.PostCode{
		OrgID:               masterLink.OrgID,
		Code:                code,
		Content:             content,
		CreatedByTelegramID: msg.From.ID,
	}

	if err := h.deps.PostCodes.Create(ctx, pc); err != nil {
		h.logger.Error("create post code", "error", err, "org_id", masterLink.OrgID)
		h.sendText(msg.Chat.ID, "Ошибка создания поста. Попробуйте позже.")
		return
	}

	h.logger.Info("post code created",
		"code", code,
		"org_id", masterLink.OrgID,
		"telegram_id", msg.From.ID,
	)

	// Send preview (forward message back)
	// Then send management message with code + buttons
	h.sendPostManagement(ctx, msg.Chat.ID, pc)
}

// extractContent extracts text and media info from a Telegram message.
func (h *handler) extractContent(msg *telego.Message) entity.PostCodeContent {
	content := entity.PostCodeContent{}

	// Extract text
	if msg.Text != "" {
		content.Text = msg.Text
	} else if msg.Caption != "" {
		content.Text = msg.Caption
	}

	// Extract media
	switch {
	case msg.Photo != nil && len(msg.Photo) > 0:
		// Use largest photo
		largest := msg.Photo[len(msg.Photo)-1]
		content.MediaURLs = []string{largest.FileID}
		content.MediaType = "photo"
	case msg.Video != nil:
		content.MediaURLs = []string{msg.Video.FileID}
		content.MediaType = "video"
	case msg.Document != nil:
		content.MediaURLs = []string{msg.Document.FileID}
		content.MediaType = "document"
	case msg.Animation != nil:
		content.MediaURLs = []string{msg.Animation.FileID}
		content.MediaType = "animation"
	case msg.Audio != nil:
		content.MediaURLs = []string{msg.Audio.FileID}
		content.MediaType = "audio"
	case msg.Voice != nil:
		content.MediaURLs = []string{msg.Voice.FileID}
		content.MediaType = "voice"
	case msg.Sticker != nil:
		content.MediaURLs = []string{msg.Sticker.FileID}
		content.MediaType = "sticker"
	}

	return content
}

func (h *handler) sendPostManagement(ctx context.Context, chatID int64, pc *entity.PostCode) {
	text := fmt.Sprintf("✅ Пост создан!\n\n"+
		"Код поста: `%s`\n\n"+
		"Используйте этот код на сайте при создании рассылки,\n"+
		"или нажмите «Скопировать код» ниже.", pc.Code)

	kb := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("📋 Скопировать код").WithCallbackData("copy_code:"+pc.Code),
		),
	)

	tgMsg := tu.Message(tu.ID(chatID), text).
		WithParseMode("Markdown").
		WithReplyMarkup(kb)

	if _, err := h.bot.SendMessage(context.Background(), tgMsg); err != nil {
		h.logger.Error("send post management", "error", err, "chat_id", chatID)
	}
}
