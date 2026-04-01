package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// Sender sends MessageContent through the Telegram Bot API.
// Used by: campaign sender, welcome message, auto-actions.
type Sender struct {
	baseURL string // Base URL for media (e.g., "https://elysium.fm")
	logger  *slog.Logger
}

func NewSender(baseURL string, logger *slog.Logger) *Sender {
	return &Sender{baseURL: baseURL, logger: logger}
}

// SendContent sends a composite message (all parts sequentially).
func (s *Sender) SendContent(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent) error {
	for i, part := range content.Parts {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Buttons attach only to the last part
		var markup *telego.InlineKeyboardMarkup
		if i == len(content.Parts)-1 && len(content.Buttons) > 0 {
			markup = s.buildInlineKeyboard(content.Buttons)
		}

		if err := s.sendPart(bot, chatID, part, markup); err != nil {
			return fmt.Errorf("send part %d (%s): %w", i, part.Type, err)
		}

		// Rate limiting: pause between parts to avoid Telegram 429 errors.
		// Telegram allows ~30 msgs/sec per bot; 50ms gap prevents bursts.
		if i < len(content.Parts)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return nil
}

func (s *Sender) sendPart(
	bot *telego.Bot,
	chatID int64,
	part entity.MessagePart,
	markup *telego.InlineKeyboardMarkup,
) error {
	switch part.Type {
	case entity.PartText:
		msg := tu.Message(tu.ID(chatID), part.Text)
		if part.ParseMode != "" {
			msg = msg.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			msg = msg.WithReplyMarkup(markup)
		}
		_, err := bot.SendMessage(msg)
		return err

	case entity.PartPhoto:
		photo := tu.Photo(tu.ID(chatID), s.mediaInput(part))
		if part.Text != "" {
			photo = photo.WithCaption(part.Text)
		}
		if part.ParseMode != "" {
			photo = photo.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			photo = photo.WithReplyMarkup(markup)
		}
		_, err := bot.SendPhoto(photo)
		return err

	case entity.PartVideo:
		video := tu.Video(tu.ID(chatID), s.mediaInput(part))
		if part.Text != "" {
			video = video.WithCaption(part.Text)
		}
		if part.ParseMode != "" {
			video = video.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			video = video.WithReplyMarkup(markup)
		}
		_, err := bot.SendVideo(video)
		return err

	case entity.PartDocument:
		doc := tu.Document(tu.ID(chatID), s.mediaInput(part))
		if part.Text != "" {
			doc = doc.WithCaption(part.Text)
		}
		if markup != nil {
			doc = doc.WithReplyMarkup(markup)
		}
		_, err := bot.SendDocument(doc)
		return err

	case entity.PartSticker:
		sticker := tu.Sticker(tu.ID(chatID), s.mediaInput(part))
		_, err := bot.SendSticker(sticker)
		return err

	case entity.PartAnimation:
		anim := tu.Animation(tu.ID(chatID), s.mediaInput(part))
		if part.Text != "" {
			anim = anim.WithCaption(part.Text)
		}
		if markup != nil {
			anim = anim.WithReplyMarkup(markup)
		}
		_, err := bot.SendAnimation(anim)
		return err

	case entity.PartAudio:
		audio := tu.Audio(tu.ID(chatID), s.mediaInput(part))
		if part.Text != "" {
			audio = audio.WithCaption(part.Text)
		}
		if markup != nil {
			audio = audio.WithReplyMarkup(markup)
		}
		_, err := bot.SendAudio(audio)
		return err

	case entity.PartVoice:
		voice := tu.Voice(tu.ID(chatID), s.mediaInput(part))
		if part.Text != "" {
			voice = voice.WithCaption(part.Text)
		}
		if markup != nil {
			voice = voice.WithReplyMarkup(markup)
		}
		_, err := bot.SendVoice(voice)
		return err

	default:
		return fmt.Errorf("unsupported part type: %s", part.Type)
	}
}

// mediaInput selects FileFromID (cached) or FileFromURL.
func (s *Sender) mediaInput(part entity.MessagePart) telego.InputFile {
	if part.MediaID != "" {
		return tu.FileFromID(part.MediaID)
	}
	url := part.MediaURL
	if !strings.HasPrefix(url, "http") {
		url = s.baseURL + url
	}
	return tu.FileFromURL(url)
}

func (s *Sender) buildInlineKeyboard(buttons [][]entity.InlineButton) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton
	for _, row := range buttons {
		var tgRow []telego.InlineKeyboardButton
		for _, btn := range row {
			if btn.URL != "" {
				tgRow = append(tgRow, tu.InlineKeyboardButton(btn.Text).WithURL(btn.URL))
			} else if btn.Data != "" {
				tgRow = append(tgRow, tu.InlineKeyboardButton(btn.Text).WithCallbackData(btn.Data))
			}
		}
		if len(tgRow) > 0 {
			rows = append(rows, tgRow)
		}
	}
	kb := tu.InlineKeyboard(rows...)
	return kb
}
