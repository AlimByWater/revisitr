package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// MessagePartType defines the type of a message part.
type MessagePartType string

const (
	PartText      MessagePartType = "text"
	PartPhoto     MessagePartType = "photo"
	PartVideo     MessagePartType = "video"
	PartDocument  MessagePartType = "document"
	PartAnimation MessagePartType = "animation" // GIF
	PartSticker   MessagePartType = "sticker"
	PartAudio     MessagePartType = "audio"
	PartVoice     MessagePartType = "voice"
)

// MessagePart is one part of a composite message.
// Each part corresponds to one Telegram API call.
type MessagePart struct {
	Type      MessagePartType `json:"type"`
	Text      string          `json:"text,omitempty"`       // Text or caption
	MediaURL  string          `json:"media_url,omitempty"`  // File URL (MinIO)
	MediaID   string          `json:"media_id,omitempty"`   // Telegram file_id (cache)
	ParseMode string          `json:"parse_mode,omitempty"` // "Markdown", "HTML", ""
}

// InlineButton is a button attached to the last message part.
type InlineButton struct {
	Text string `json:"text"`
	URL  string `json:"url,omitempty"`
	Data string `json:"data,omitempty"` // callback_data
}

// MessageContent is a full composite message description.
// Stored as JSONB in PostgreSQL.
type MessageContent struct {
	Parts   []MessagePart    `json:"parts"`
	Buttons [][]InlineButton `json:"buttons,omitempty"` // rows of buttons
}

func (mc *MessageContent) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, mc)
	case string:
		return json.Unmarshal([]byte(v), mc)
	case nil:
		*mc = MessageContent{}
		return nil
	default:
		return fmt.Errorf("MessageContent.Scan: unsupported type %T", src)
	}
}

func (mc MessageContent) Value() (driver.Value, error) {
	b, err := json.Marshal(mc)
	if err != nil {
		return nil, fmt.Errorf("MessageContent.Value: %w", err)
	}
	return b, nil
}

// Validate checks the message content for correctness.
func (mc MessageContent) Validate() error {
	if len(mc.Parts) == 0 {
		return errors.New("message must have at least one part")
	}
	if len(mc.Parts) > 5 {
		return errors.New("message cannot have more than 5 parts")
	}

	for i, p := range mc.Parts {
		switch p.Type {
		case PartText:
			if p.Text == "" {
				return fmt.Errorf("part %d: text part must have text", i)
			}
			if p.MediaURL != "" {
				return fmt.Errorf("part %d: text part cannot have media", i)
			}
		case PartPhoto, PartVideo, PartDocument, PartAnimation, PartAudio, PartVoice:
			if p.MediaURL == "" && p.MediaID == "" {
				return fmt.Errorf("part %d: media part must have media_url or media_id", i)
			}
		case PartSticker:
			if p.MediaURL == "" && p.MediaID == "" {
				return fmt.Errorf("part %d: sticker must have media_url or media_id", i)
			}
			if p.Text != "" {
				return fmt.Errorf("part %d: stickers cannot have captions", i)
			}
		default:
			return fmt.Errorf("part %d: unknown type %q", i, p.Type)
		}
	}

	for _, row := range mc.Buttons {
		if len(row) > 8 {
			return errors.New("button row cannot have more than 8 buttons")
		}
		for _, btn := range row {
			if btn.Text == "" {
				return errors.New("button must have text")
			}
		}
	}

	return nil
}

// TextContent creates a simple text-only MessageContent (convenience helper).
func TextContent(text, parseMode string) MessageContent {
	return MessageContent{
		Parts: []MessagePart{
			{Type: PartText, Text: text, ParseMode: parseMode},
		},
	}
}
