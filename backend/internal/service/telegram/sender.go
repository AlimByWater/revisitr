package telegram

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
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

		if err := s.sendPart(ctx, bot, chatID, part, markup); err != nil {
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
	ctx context.Context,
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
		_, err := bot.SendMessage(ctx, msg)
		return err

	case entity.PartPhoto:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		photo := tu.Photo(tu.ID(chatID), media)
		if part.Text != "" {
			photo = photo.WithCaption(part.Text)
		}
		if part.ParseMode != "" {
			photo = photo.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			photo = photo.WithReplyMarkup(markup)
		}
		_, err = bot.SendPhoto(ctx, photo)
		return err

	case entity.PartVideo:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		video := tu.Video(tu.ID(chatID), media)
		if part.Text != "" {
			video = video.WithCaption(part.Text)
		}
		if part.ParseMode != "" {
			video = video.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			video = video.WithReplyMarkup(markup)
		}
		_, err = bot.SendVideo(ctx, video)
		return err

	case entity.PartDocument:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		doc := tu.Document(tu.ID(chatID), media)
		if part.Text != "" {
			doc = doc.WithCaption(part.Text)
		}
		if markup != nil {
			doc = doc.WithReplyMarkup(markup)
		}
		_, err = bot.SendDocument(ctx, doc)
		return err

	case entity.PartSticker:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		sticker := tu.Sticker(tu.ID(chatID), media)
		_, err = bot.SendSticker(ctx, sticker)
		return err

	case entity.PartAnimation:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		anim := tu.Animation(tu.ID(chatID), media)
		if part.Text != "" {
			anim = anim.WithCaption(part.Text)
		}
		if markup != nil {
			anim = anim.WithReplyMarkup(markup)
		}
		_, err = bot.SendAnimation(ctx, anim)
		return err

	case entity.PartAudio:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		audio := tu.Audio(tu.ID(chatID), media)
		if part.Text != "" {
			audio = audio.WithCaption(part.Text)
		}
		if markup != nil {
			audio = audio.WithReplyMarkup(markup)
		}
		_, err = bot.SendAudio(ctx, audio)
		return err

	case entity.PartVoice:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return err
		}
		voice := tu.Voice(tu.ID(chatID), media)
		if part.Text != "" {
			voice = voice.WithCaption(part.Text)
		}
		if markup != nil {
			voice = voice.WithReplyMarkup(markup)
		}
		_, err = bot.SendVoice(ctx, voice)
		return err

	default:
		return fmt.Errorf("unsupported part type: %s", part.Type)
	}
}

// mediaInput selects FileFromID (cached), or downloads bytes from a public URL
// so uploads don't depend on Telegram fetching external URLs itself.
func (s *Sender) mediaInput(ctx context.Context, part entity.MessagePart) (telego.InputFile, error) {
	if part.MediaID != "" {
		return tu.FileFromID(part.MediaID), nil
	}
	url := part.MediaURL
	if !strings.HasPrefix(url, "http") {
		url = s.baseURL + url
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return telego.InputFile{}, fmt.Errorf("build media request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return telego.InputFile{}, fmt.Errorf("download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return telego.InputFile{}, fmt.Errorf("download media: unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 25<<20))
	if err != nil {
		return telego.InputFile{}, fmt.Errorf("read media body: %w", err)
	}

	filename := path.Base(part.MediaURL)
	if filename == "." || filename == "/" || filename == "" {
		filename = "media"
	}

	return tu.FileFromBytes(data, filename), nil
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
