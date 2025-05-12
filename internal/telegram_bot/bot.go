package telegram_bot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBot interface {
	Init(onMessage func(ctx context.Context, update *models.Update)) error
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type botHandler struct {
	instance *bot.Bot
}

func newBotHandler() TelegramBot {
	return &botHandler{}
}

func (h *botHandler) Init(onMessage func(ctx context.Context, update *models.Update)) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			onMessage(ctx, update)
		}),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, h.startCommand),
		bot.WithMessageTextHandler("/help", bot.MatchTypeExact, h.helpCommand),
	}

	var err error
	h.instance, err = bot.New(token, opts...)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	log.Println("🤖 Starting Telegram bot...")

	// h.SendMessage(context.TODO(), 574037714, "Bot started")

	h.instance.Start(context.TODO())

	return nil
}

func (h *botHandler) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := h.instance.SendMessage(context.TODO(), &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	return err
}

func (h *botHandler) startCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendWelcomeMessage(ctx, update.Message.Chat.ID)
}

func (h *botHandler) helpCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.SendMessage(ctx, update.Message.Chat.ID, helpMessage)
}

func (h *botHandler) sendWelcomeMessage(ctx context.Context, chatID int64) {
	h.SendMessage(ctx, chatID, startMessage)
}
