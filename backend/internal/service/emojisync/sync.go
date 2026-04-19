package emojisync

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"revisitr/internal/entity"
)

const stickerSize = 100

type namedReader struct {
	io.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

// Service syncs emoji pack items to Telegram custom emoji sticker sets.
type Service struct {
	logger  *slog.Logger
	baseURL string // e.g. "https://elysium.fm"
}

func New(logger *slog.Logger, baseURL string) *Service {
	return &Service{logger: logger, baseURL: strings.TrimRight(baseURL, "/")}
}

// SyncPack creates or updates a Telegram custom emoji sticker set for the given pack.
func (s *Service) SyncPack(ctx context.Context, bot *telego.Bot, botUsername string, ownerTgID int64, pack *entity.EmojiPack) ([]entity.EmojiItem, error) {
	setName := stickerSetName(pack.ID, botUsername)
	setTitle := pack.Name

	// Check if sticker set already exists
	existing, err := bot.GetStickerSet(ctx, &telego.GetStickerSetParams{Name: setName})
	setExists := err == nil && existing != nil

	var results []entity.EmojiItem
	for i, item := range pack.Items {
		if item.TgCustomEmojiID != nil && *item.TgCustomEmojiID != "" {
			results = append(results, item)
			continue
		}

		stickerData, err := s.prepareSticker(ctx, item.ImageURL)
		if err != nil {
			s.logger.Error("prepare sticker", "error", err, "item_id", item.ID)
			results = append(results, item)
			continue
		}

		stickerFile := tu.File(namedReader{
			Reader: bytes.NewReader(stickerData),
			name:   fmt.Sprintf("emoji_%d.png", item.ID),
		})

		inputSticker := telego.InputSticker{
			Sticker:   stickerFile,
			Format:    "static",
			EmojiList: []string{"⭐"},
		}

		if !setExists && i == 0 {
			err = bot.CreateNewStickerSet(ctx, &telego.CreateNewStickerSetParams{
				UserID:      ownerTgID,
				Name:        setName,
				Title:       setTitle,
				Stickers:    []telego.InputSticker{inputSticker},
				StickerType: "custom_emoji",
			})
			if err != nil {
				return nil, fmt.Errorf("create sticker set: %w", err)
			}
			setExists = true
		} else {
			err = bot.AddStickerToSet(ctx, &telego.AddStickerToSetParams{
				UserID:  ownerTgID,
				Name:    setName,
				Sticker: inputSticker,
			})
			if err != nil {
				s.logger.Error("add sticker to set", "error", err, "item_id", item.ID)
				results = append(results, item)
				continue
			}
		}

		// Fetch updated set to get custom_emoji_id
		updatedSet, err := bot.GetStickerSet(ctx, &telego.GetStickerSetParams{Name: setName})
		if err != nil {
			s.logger.Error("get sticker set after add", "error", err)
			results = append(results, item)
			continue
		}

		if len(updatedSet.Stickers) > 0 {
			lastSticker := updatedSet.Stickers[len(updatedSet.Stickers)-1]
			item.TgStickerSet = &setName
			item.TgCustomEmojiID = &lastSticker.CustomEmojiID
		}

		results = append(results, item)
	}

	return results, nil
}

func (s *Service) resolveURL(imageURL string) string {
	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		return imageURL
	}
	return s.baseURL + imageURL
}

func (s *Service) prepareSticker(ctx context.Context, imageURL string) ([]byte, error) {
	fullURL := s.resolveURL(imageURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download image: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read image body: %w", err)
	}

	img, err := imaging.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	resized := imaging.Fill(img, stickerSize, stickerSize, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}

	return buf.Bytes(), nil
}

func stickerSetName(packID int, botUsername string) string {
	return fmt.Sprintf("emoji_pack_%d_by_%s", packID, strings.ToLower(botUsername))
}
