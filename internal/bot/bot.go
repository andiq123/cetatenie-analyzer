package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/andiq123/cetatenie-analyzer/internal/decree_processor"
	"github.com/andiq123/cetatenie-analyzer/internal/parser"
	"github.com/andiq123/cetatenie-analyzer/internal/timer"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// BotHandler handles Telegram bot interactions
type BotHandler struct {
	decreeProcessor decree_processor.Processor
	botInstance     *bot.Bot
}

// New creates a new BotHandler instance
func New(dp decree_processor.Processor) *BotHandler {
	return &BotHandler{
		decreeProcessor: dp,
	}
}

// Init initializes and starts the Telegram bot
func (h *BotHandler) Init() error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(h.defaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, h.startCommand),
		bot.WithMessageTextHandler("/help", bot.MatchTypeExact, h.helpCommand),
	}

	botInstance, err := bot.New(token, opts...)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	h.botInstance = botInstance

	log.Println("🤖 Starting Telegram bot...")
	botInstance.Start(context.TODO())
	log.Println("🛑 Telegram bot stopped")

	return nil
}

// Close cleans up resources
func (h *BotHandler) Close() error {
	if h.decreeProcessor != nil {
		return h.decreeProcessor.CleanUpCache()
	}
	return nil
}

func (h *BotHandler) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if regexp.MustCompile(decreePattern).MatchString(update.Message.Text) {
		h.handleDecreeRequest(ctx, b, update)
		return
	}

	// Send invalid format message if the message doesn't match any command
	h.sendMessage(ctx, b, update.Message.Chat.ID, invalidFormat)
}

func (h *BotHandler) startCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendWelcomeMessage(ctx, b, update.Message.Chat.ID)
}

func (h *BotHandler) helpCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendMessage(ctx, b, update.Message.Chat.ID, helpMessage)
}

func (h *BotHandler) handleDecreeRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	decreeNumber := strings.TrimSpace(update.Message.Text)

	if err := h.sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf(searching, decreeNumber)); err != nil {
		log.Printf("Error sending searching message: %v", err)
		return
	}

	findState, timeReport, err := h.decreeProcessor.Handle(decreeNumber)
	if err != nil {
		h.sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf(errorMessage, err.Error()))
		return
	}

	var response string
	switch findState {
	case parser.StateFoundAndResolved:
		response = fmt.Sprintf(successMessage, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	case parser.StateFoundButNotResolved:
		response = fmt.Sprintf(inProgressMsg, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	case parser.StateNotFound:
		response = fmt.Sprintf(notFoundMsg, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	default:
		response = unknownState
	}

	if err := h.sendMessage(ctx, b, update.Message.Chat.ID, response); err != nil {
		log.Printf("Error sending response message: %v", err)
	}
}

func (h *BotHandler) sendWelcomeMessage(ctx context.Context, b *bot.Bot, chatID int64) {
	h.sendMessage(ctx, b, chatID, startMessage)
}

func (h *BotHandler) sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	return err
}
