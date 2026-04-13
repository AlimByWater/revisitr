package config

import (
	"time"

	"revisitr/internal/application/env"
)

type Http struct {
	Port string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type Redis struct {
	Host     string
	Port     string
	Password string
}

type Auth struct {
	JWTSecret string
	TokenTTL  time.Duration
}

type Bot struct {
	Token string
}

type AdminBot struct {
	Token string
}

type MinIO struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

type Module struct {
	Http     Http
	Postgres Postgres
	Redis    Redis
	Auth     Auth
	Bot      Bot
	AdminBot AdminBot
	MinIO    MinIO
	BaseURL        string // Public base URL for media files (e.g., "https://elysium.fm")
	TelegramAPIURL string // Custom Telegram Bot API server URL (empty = default api.telegram.org)
}

func (m *Module) GetBaseURL() string {
	return m.BaseURL
}

func NewFromEnv() *Module {
	ttlMinutes := env.GetInt("AUTH_TOKEN_TTL_MINUTES", 60)

	return &Module{
		Http: Http{
			Port: env.GetString("HTTP_PORT", "8080"),
		},
		Postgres: Postgres{
			Host:     env.GetString("POSTGRES_HOST", "localhost"),
			Port:     env.GetString("POSTGRES_PORT", "5432"),
			User:     env.GetString("POSTGRES_USER", "revisitr"),
			Password: env.GetString("POSTGRES_PASSWORD", "revisitr"),
			Database: env.GetString("POSTGRES_DATABASE", "revisitr"),
			SSLMode:  env.GetString("POSTGRES_SSLMODE", "disable"),
		},
		Redis: Redis{
			Host:     env.GetString("REDIS_HOST", "localhost"),
			Port:     env.GetString("REDIS_PORT", "6379"),
			Password: env.GetString("REDIS_PASSWORD", ""),
		},
		Auth: Auth{
			JWTSecret: env.GetString("AUTH_JWT_SECRET", "change-me-in-production"),
			TokenTTL:  time.Duration(ttlMinutes) * time.Minute,
		},
		Bot: Bot{
			Token: env.GetString("BOT_TOKEN", ""),
		},
		AdminBot: AdminBot{
			Token: env.GetString("ADMIN_BOT_TOKEN", ""),
		},
		MinIO: MinIO{
			Endpoint:  env.GetString("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: env.GetString("MINIO_ACCESS_KEY", "revisitr"),
			SecretKey: env.GetString("MINIO_SECRET_KEY", "devpassword"),
			UseSSL:    env.GetBool("MINIO_USE_SSL", false),
			Bucket:    env.GetString("MINIO_BUCKET", "revisitr"),
		},
		BaseURL:        env.GetString("BASE_URL", "https://elysium.fm"),
		TelegramAPIURL: env.GetString("TELEGRAM_API_URL", ""),
	}
}
