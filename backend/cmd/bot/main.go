package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"revisitr/internal/application/config"
	"revisitr/internal/application/env"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	envModule := &env.Module{}
	if err := envModule.Init(); err != nil {
		logger.Warn("no .env file loaded, using environment variables", "error", err)
	}

	cfg := config.NewFromEnv()

	if cfg.Bot.Token == "" {
		logger.Error("BOT_TOKEN is required")
		os.Exit(1)
	}

	bot, err := telego.NewBot(cfg.Bot.Token)
	if err != nil {
		logger.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		logger.Error("failed to start long polling", "error", err)
		os.Exit(1)
	}

	logger.Info("bot started, waiting for updates")

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.Text == "/start" {
				msg := tu.Message(
					tu.ID(update.Message.Chat.ID),
					fmt.Sprintf("Добро пожаловать в Revisitr! 🎉"),
				)
				if _, err := bot.SendMessage(msg); err != nil {
					logger.Error("failed to send message", "error", err)
				}
				logger.Info("handled /start",
					"chat_id", update.Message.Chat.ID,
					"user", update.Message.From.Username,
				)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down bot")
	bot.StopLongPolling()
	logger.Info("bot stopped")
}
