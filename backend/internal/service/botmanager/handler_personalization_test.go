package botmanager

import (
	"encoding/json"
	"testing"
	"time"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
)

func TestPersonalizeMessageContentUsesClientValues(t *testing.T) {
	birthDate := time.Date(1991, 4, 23, 0, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"favorite_drink":"чай"}`)
	client := &entity.BotClient{
		FirstName: "Алексей",
		LastName:  "Иванов",
		Username:  "alex",
		Phone:     "+79990000000",
		BirthDate: &birthDate,
		Data:      data,
	}

	content := entity.MessageContent{
		Parts: []entity.MessagePart{
			{Type: entity.PartText, Text: "Привет, {first_name}! Напиток: {favorite_drink}. ДР: {birth_date}."},
			{Type: entity.PartPhoto, Text: "Фото для {name}"},
		},
		Buttons: [][]entity.InlineButton{{{Text: "{first_name}, открыть"}}},
	}

	got := personalizeMessageContent(content, templateValues(client, nil))

	if got.Parts[0].Text != "Привет, Алексей! Напиток: чай. ДР: 1991-04-23." {
		t.Fatalf("text part = %q", got.Parts[0].Text)
	}
	if got.Parts[1].Text != "Фото для Алексей" {
		t.Fatalf("caption = %q", got.Parts[1].Text)
	}
	if got.Buttons[0][0].Text != "Алексей, открыть" {
		t.Fatalf("button = %q", got.Buttons[0][0].Text)
	}
}

func TestTemplateValuesFallBackToTelegramUser(t *testing.T) {
	values := templateValues(nil, &telego.User{FirstName: "Мария", LastName: "Петрова", Username: "maria"})

	if got := personalizeText("Здравствуйте, {first_name} {last_name} (@{username})", values); got != "Здравствуйте, Мария Петрова (@maria)" {
		t.Fatalf("personalized = %q", got)
	}
}

func TestMergeMediaIDsKeepsOriginalTemplateText(t *testing.T) {
	base := entity.MessageContent{
		Parts: []entity.MessagePart{{Type: entity.PartPhoto, Text: "Привет, {first_name}", MediaURL: "/file.jpg"}},
	}
	sent := entity.MessageContent{
		Parts: []entity.MessagePart{{Type: entity.PartPhoto, Text: "Привет, Алексей", MediaURL: "/file.jpg", MediaID: "photo-file-id"}},
	}

	got := mergeMediaIDs(base, sent)

	if got.Parts[0].Text != "Привет, {first_name}" {
		t.Fatalf("template text was changed: %q", got.Parts[0].Text)
	}
	if got.Parts[0].MediaID != "photo-file-id" {
		t.Fatalf("media id = %q", got.Parts[0].MediaID)
	}
}
