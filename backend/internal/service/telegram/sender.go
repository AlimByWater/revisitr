package telegram

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// EmojiResolver maps image URLs to Telegram custom_emoji_ids.
type EmojiResolver interface {
	GetSyncedItemsByOrgID(ctx context.Context, orgID int) ([]entity.EmojiItem, error)
}

// Sender sends MessageContent through the Telegram Bot API.
// Used by: campaign sender, welcome message, auto-actions.
type Sender struct {
	baseURL       string
	logger        *slog.Logger
	emojiResolver EmojiResolver
}

func NewSender(baseURL string, logger *slog.Logger) *Sender {
	return &Sender{baseURL: baseURL, logger: logger}
}

func (s *Sender) WithEmojiResolver(r EmojiResolver) *Sender {
	s.emojiResolver = r
	return s
}

var emojiMarkerRe = regexp.MustCompile(`\{\{emoji:([^}]+)\}\}`)

// resolveEmoji builds URL→custom_emoji_id map for an org.
func (s *Sender) resolveEmoji(ctx context.Context, orgID int) map[string]string {
	if s.emojiResolver == nil {
		return nil
	}
	items, err := s.emojiResolver.GetSyncedItemsByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("resolve emoji", "error", err)
		return nil
	}
	m := make(map[string]string, len(items))
	for _, item := range items {
		if item.TgCustomEmojiID != nil {
			m[item.ImageURL] = *item.TgCustomEmojiID
		}
	}
	return m
}

// processEmojiText replaces {{emoji:URL}} markers with ⭐ placeholder and builds entities.
func processEmojiText(text string, emojiMap map[string]string) (string, []telego.MessageEntity) {
	if !strings.Contains(text, "{{emoji:") || len(emojiMap) == 0 {
		// Strip markers if no emoji map
		cleaned := emojiMarkerRe.ReplaceAllString(text, "")
		return strings.TrimSpace(cleaned), nil
	}

	var result strings.Builder
	var entities []telego.MessageEntity
	lastIndex := 0

	for _, loc := range emojiMarkerRe.FindAllStringSubmatchIndex(text, -1) {
		// loc[0:1] = full match, loc[2:3] = capture group
		result.WriteString(text[lastIndex:loc[0]])
		imageURL := text[loc[2]:loc[3]]

		customEmojiID, ok := emojiMap[imageURL]
		if ok {
			offset := utf16Len(result.String())
			result.WriteString("⭐")
			entities = append(entities, telego.MessageEntity{
				Type:          "custom_emoji",
				Offset:        offset,
				Length:        utf16Len("⭐"),
				CustomEmojiID: customEmojiID,
			})
		}
		// If not synced, skip (remove marker)
		lastIndex = loc[1]
	}
	result.WriteString(text[lastIndex:])

	return strings.TrimSpace(result.String()), entities
}

// utf16Len returns length of string in UTF-16 code units (Telegram offset unit).
func utf16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// SendContentOpts controls optional behavior when sending composite messages.
type SendContentOpts struct {
	// ReplyKeyboard attaches a reply keyboard to the last message part.
	// When set, inline buttons from MessageContent.Buttons are ignored
	// (Telegram API accepts only one reply_markup per message).
	ReplyKeyboard *telego.ReplyKeyboardMarkup
}

// SendContent sends a composite message (all parts sequentially).
func (s *Sender) SendContent(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent) error {
	_, _, err := s.SendContentWithCache(ctx, bot, chatID, content)
	return err
}

// SendContentForOrg sends with emoji resolution for the given org.
func (s *Sender) SendContentForOrg(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent, orgID int) (entity.MessageContent, bool, error) {
	emojiMap := s.resolveEmoji(ctx, orgID)
	return s.sendContentInternal(ctx, bot, chatID, content, emojiMap, nil)
}

// SendContentForOrgWithOpts sends with emoji resolution and additional options (e.g. reply keyboard).
func (s *Sender) SendContentForOrgWithOpts(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent, orgID int, opts *SendContentOpts) (entity.MessageContent, bool, error) {
	emojiMap := s.resolveEmoji(ctx, orgID)
	return s.sendContentInternal(ctx, bot, chatID, content, emojiMap, opts)
}

// SendContentWithCache sends a composite message and returns updated content
// with resolved Telegram media_id values when they were learned during send.
func (s *Sender) SendContentWithCache(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent) (entity.MessageContent, bool, error) {
	return s.sendContentInternal(ctx, bot, chatID, content, nil, nil)
}

func (s *Sender) sendContentInternal(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent, emojiMap map[string]string, opts *SendContentOpts) (entity.MessageContent, bool, error) {
	updated := content
	changed := false

	for i, part := range updated.Parts {
		if ctx.Err() != nil {
			return updated, changed, ctx.Err()
		}

		// Determine markup for the last part only.
		// Reply keyboard takes priority (mutually exclusive with inline buttons in Telegram API).
		var lastPartMarkup telego.ReplyMarkup
		if i == len(updated.Parts)-1 {
			if opts != nil && opts.ReplyKeyboard != nil {
				lastPartMarkup = opts.ReplyKeyboard
			} else if len(updated.Buttons) > 0 {
				lastPartMarkup = s.buildInlineKeyboard(updated.Buttons)
			}
		}

		mediaID, err := s.sendPart(ctx, bot, chatID, part, lastPartMarkup, emojiMap)
		if err != nil {
			return updated, changed, fmt.Errorf("send part %d (%s): %w", i, part.Type, err)
		}
		if mediaID != "" && updated.Parts[i].MediaID != mediaID {
			updated.Parts[i].MediaID = mediaID
			changed = true
		}

		if i < len(updated.Parts)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return updated, changed, nil
}

func (s *Sender) sendPart(
	ctx context.Context,
	bot *telego.Bot,
	chatID int64,
	part entity.MessagePart,
	markup telego.ReplyMarkup,
	emojiMap map[string]string,
) (string, error) {
	switch part.Type {
	case entity.PartText:
		cleanText, emojiEntities := processEmojiText(part.Text, emojiMap)
		if cleanText == "" && len(emojiEntities) == 0 {
			return "", nil // skip empty text part
		}
		msg := tu.Message(tu.ID(chatID), cleanText)
		if len(emojiEntities) > 0 {
			msg = msg.WithEntities(emojiEntities...)
		} else if part.ParseMode != "" {
			msg = msg.WithParseMode(part.ParseMode)
		}
		if markup != nil {
			msg = msg.WithReplyMarkup(markup)
		}
		_, err := bot.SendMessage(ctx, msg)
		return "", err

	case entity.PartPhoto:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		photo := tu.Photo(tu.ID(chatID), media)
		if part.Text != "" {
			cleanCaption, captionEntities := processEmojiText(part.Text, emojiMap)
			photo = photo.WithCaption(cleanCaption)
			if len(captionEntities) > 0 {
				photo = photo.WithCaptionEntities(captionEntities...)
			} else if part.ParseMode != "" {
				photo = photo.WithParseMode(part.ParseMode)
			}
		}
		if markup != nil {
			photo = photo.WithReplyMarkup(markup)
		}
		message, err := bot.SendPhoto(ctx, photo)
		return s.extractMediaID(message, part.Type), err

	case entity.PartVideo:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		video := tu.Video(tu.ID(chatID), media)
		if part.Text != "" {
			cleanCaption, captionEntities := processEmojiText(part.Text, emojiMap)
			video = video.WithCaption(cleanCaption)
			if len(captionEntities) > 0 {
				video = video.WithCaptionEntities(captionEntities...)
			} else if part.ParseMode != "" {
				video = video.WithParseMode(part.ParseMode)
			}
		}
		if markup != nil {
			video = video.WithReplyMarkup(markup)
		}
		message, err := bot.SendVideo(ctx, video)
		return s.extractMediaID(message, part.Type), err

	case entity.PartDocument:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		doc := tu.Document(tu.ID(chatID), media)
		if part.Text != "" {
			cleanCaption, _ := processEmojiText(part.Text, emojiMap)
			doc = doc.WithCaption(cleanCaption)
		}
		if markup != nil {
			doc = doc.WithReplyMarkup(markup)
		}
		message, err := bot.SendDocument(ctx, doc)
		return s.extractMediaID(message, part.Type), err

	case entity.PartSticker:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		sticker := tu.Sticker(tu.ID(chatID), media)
		if markup != nil {
			sticker = sticker.WithReplyMarkup(markup)
		}
		message, err := bot.SendSticker(ctx, sticker)
		return s.extractMediaID(message, part.Type), err

	case entity.PartAnimation:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		anim := tu.Animation(tu.ID(chatID), media)
		if part.Text != "" {
			cleanCaption, _ := processEmojiText(part.Text, emojiMap)
			anim = anim.WithCaption(cleanCaption)
		}
		if markup != nil {
			anim = anim.WithReplyMarkup(markup)
		}
		message, err := bot.SendAnimation(ctx, anim)
		return s.extractMediaID(message, part.Type), err

	case entity.PartAudio:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		audio := tu.Audio(tu.ID(chatID), media)
		if part.Text != "" {
			cleanCaption, _ := processEmojiText(part.Text, emojiMap)
			audio = audio.WithCaption(cleanCaption)
		}
		if markup != nil {
			audio = audio.WithReplyMarkup(markup)
		}
		message, err := bot.SendAudio(ctx, audio)
		return s.extractMediaID(message, part.Type), err

	case entity.PartVoice:
		media, err := s.mediaInput(ctx, part)
		if err != nil {
			return "", err
		}
		voice := tu.Voice(tu.ID(chatID), media)
		if part.Text != "" {
			cleanCaption, _ := processEmojiText(part.Text, emojiMap)
			voice = voice.WithCaption(cleanCaption)
		}
		if markup != nil {
			voice = voice.WithReplyMarkup(markup)
		}
		message, err := bot.SendVoice(ctx, voice)
		return s.extractMediaID(message, part.Type), err

	default:
		return "", fmt.Errorf("unsupported part type: %s", part.Type)
	}
}

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

func (s *Sender) extractMediaID(message *telego.Message, partType entity.MessagePartType) string {
	if message == nil {
		return ""
	}

	switch partType {
	case entity.PartPhoto:
		if len(message.Photo) == 0 {
			return ""
		}
		return message.Photo[len(message.Photo)-1].FileID
	case entity.PartVideo:
		if message.Video == nil {
			return ""
		}
		return message.Video.FileID
	case entity.PartDocument:
		if message.Document == nil {
			return ""
		}
		return message.Document.FileID
	case entity.PartSticker:
		if message.Sticker == nil {
			return ""
		}
		return message.Sticker.FileID
	case entity.PartAnimation:
		if message.Animation == nil {
			return ""
		}
		return message.Animation.FileID
	case entity.PartAudio:
		if message.Audio == nil {
			return ""
		}
		return message.Audio.FileID
	case entity.PartVoice:
		if message.Voice == nil {
			return ""
		}
		return message.Voice.FileID
	default:
		return ""
	}
}

func (s *Sender) buildInlineKeyboard(buttons [][]entity.InlineButton) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton
	for _, row := range buttons {
		var tgRow []telego.InlineKeyboardButton
		for _, btn := range row {
			var tgBtn telego.InlineKeyboardButton
			if btn.URL != "" {
				tgBtn = tu.InlineKeyboardButton(btn.Text).WithURL(btn.URL)
			} else if btn.Data != "" {
				tgBtn = tu.InlineKeyboardButton(btn.Text).WithCallbackData(btn.Data)
			} else {
				continue
			}
			if btn.Style != "" {
				tgBtn.Style = btn.Style
			}
			if btn.IconCustomEmojiID != "" {
				tgBtn.IconCustomEmojiID = btn.IconCustomEmojiID
			}
			tgRow = append(tgRow, tgBtn)
		}
		if len(tgRow) > 0 {
			rows = append(rows, tgRow)
		}
	}
	kb := tu.InlineKeyboard(rows...)
	return kb
}
