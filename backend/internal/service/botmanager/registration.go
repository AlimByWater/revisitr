package botmanager

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var (
	phoneRe = regexp.MustCompile(`^\+?\d[\d\s\-()]{6,18}\d$`)
	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	dateRe  = regexp.MustCompile(`^(\d{2})\.(\d{2})\.(\d{4})$`)
)

// validateRegistrationInput validates and normalizes user input for a registration field.
func validateRegistrationInput(text string, field entity.FormField) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("Пожалуйста, введите значение.")
	}

	switch field.Type {
	case "phone":
		cleaned := strings.ReplaceAll(text, " ", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		cleaned = strings.ReplaceAll(cleaned, "(", "")
		cleaned = strings.ReplaceAll(cleaned, ")", "")
		if !phoneRe.MatchString(text) {
			return "", fmt.Errorf("Введите номер телефона в формате: +7XXXXXXXXXX")
		}
		return cleaned, nil

	case "email":
		if !emailRe.MatchString(text) {
			return "", fmt.Errorf("Введите корректный email адрес.")
		}
		return strings.ToLower(text), nil

	case "date":
		if !dateRe.MatchString(text) {
			return "", fmt.Errorf("Введите дату в формате ДД.ММ.ГГГГ")
		}
		// Parse and validate the date
		_, err := parseBirthday(text)
		if err != nil {
			return "", fmt.Errorf("Некорректная дата. Формат: ДД.ММ.ГГГГ")
		}
		return text, nil

	case "choice":
		// Validate against options if defined
		if len(field.Options) > 0 {
			for _, opt := range field.Options {
				if strings.EqualFold(text, opt) {
					return opt, nil
				}
			}
			return "", fmt.Errorf("Выберите один из вариантов.")
		}
		return text, nil

	default: // "text"
		if len(text) > 200 {
			return "", fmt.Errorf("Слишком длинный текст (макс. 200 символов).")
		}
		return text, nil
	}
}

// parseBirthday parses DD.MM.YYYY format into time.Time.
func parseBirthday(s string) (time.Time, error) {
	m := dateRe.FindStringSubmatch(s)
	if m == nil {
		return time.Time{}, fmt.Errorf("invalid date format")
	}
	t, err := time.Parse("02.01.2006", s)
	if err != nil {
		return time.Time{}, err
	}
	// Sanity check: not in the future, not before 1900
	if t.After(time.Now()) || t.Year() < 1900 {
		return time.Time{}, fmt.Errorf("date out of range")
	}
	return t, nil
}

// buildChoiceKeyboard builds a reply keyboard for choice-type registration fields.
func buildChoiceKeyboard(options []string) *telego.ReplyKeyboardMarkup {
	var rows [][]telego.KeyboardButton
	for _, opt := range options {
		rows = append(rows, []telego.KeyboardButton{tu.KeyboardButton(opt)})
	}
	return &telego.ReplyKeyboardMarkup{
		Keyboard:        rows,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}
