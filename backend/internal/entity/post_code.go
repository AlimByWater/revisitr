package entity

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type PostCode struct {
	ID                  int              `db:"id" json:"id"`
	OrgID               int              `db:"org_id" json:"org_id"`
	Code                string           `db:"code" json:"code"`
	Content             PostCodeContent  `db:"content" json:"content"`
	TelegramMessageIDs  json.RawMessage  `db:"telegram_message_ids" json:"telegram_message_ids,omitempty"`
	CreatedByTelegramID int64            `db:"created_by_telegram_id" json:"created_by_telegram_id"`
	CreatedAt           time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time        `db:"updated_at" json:"updated_at"`
}

type PostCodeContent struct {
	Text      string           `json:"text,omitempty"`
	MediaURLs []string         `json:"media_urls,omitempty"`
	MediaType string           `json:"media_type,omitempty"` // "photo", "video", "document", etc.
	Buttons   [][]InlineButton `json:"buttons,omitempty"`
}

func (c *PostCodeContent) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = PostCodeContent{}
		return nil
	default:
		return fmt.Errorf("PostCodeContent.Scan: unsupported type %T", src)
	}
}

func (c PostCodeContent) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("PostCodeContent.Value: %w", err)
	}
	return b, nil
}

// GeneratePostCode returns a code like "RV-A1B2C3".
func GeneratePostCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I confusion
	var sb strings.Builder
	sb.WriteString("RV-")
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		sb.WriteByte(charset[n.Int64()])
	}
	return sb.String()
}
