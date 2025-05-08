package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/andiq123/cetatenie-analyzer/internal/decree_processor"
	"github.com/andiq123/cetatenie-analyzer/internal/parser"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	decreePattern = `^\d{1,5}/RD/\d{4}$`
	startMessage  = `Bun venit la Cetățenie Analyzer! 🇷🇴

Cu acest bot poți verifica starea dosarului tău de redobândire a cetățeniei române.

Trimite numărul dosarului în formatul: [număr]/RD/[an]
Exemplu: 123/RD/2023

Succes!`
	invalidFormat = "Format invalid. Te rog folosește formatul: [număr]/RD/[an], de exemplu: 123/RD/2023"
	searching     = "Caut dosarul: %s"
	errorMessage  = "A apărut o eroare: %s"
	unknownState  = "Stare necunoscută. Te rog încearcă mai târziu."
)

type BotHandler struct {
	decreeProcessor decree_processor.Processor
	botInstance     *bot.Bot
}

func Init() {
	handler := &BotHandler{
		decreeProcessor: decree_processor.New(),
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		panic("TELEGRAM_BOT_TOKEN is not set")
	}

	b, err := bot.New(token,
		bot.WithDefaultHandler(handler.defaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, handler.startCommand),
		bot.WithMessageTextHandler("/help", bot.MatchTypeExact, handler.helpCommand),
	)
	if err != nil {
		panic(err)
	}

	handler.botInstance = b

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Set up graceful shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("Stopping bot gracefully...")
		cancel()

		// Give some time for cleanup
		time.Sleep(1 * time.Second)
		log.Println("Bot stopped")
		os.Exit(0)
	}()

	log.Println("Bot started successfully")
	b.Start(ctx)
}

func (h *BotHandler) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	// Check if message is a decree number
	re := regexp.MustCompile(decreePattern)
	if re.MatchString(update.Message.Text) {
		h.handleDecreeRequest(ctx, b, update)
		return
	}

	// If not a command or decree number, show help
	h.sendHelpMessage(ctx, b, update.Message.Chat.ID)
}

func (h *BotHandler) startCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendMessage(ctx, b, update.Message.Chat.ID, startMessage)
}

func (h *BotHandler) helpCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendHelpMessage(ctx, b, update.Message.Chat.ID)
}

func (h *BotHandler) handleDecreeRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	decreeNumber := strings.TrimSpace(update.Message.Text)
	h.sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf(searching, decreeNumber))

	findState, err := h.decreeProcessor.Handle(decreeNumber)
	if err != nil {
		h.sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf(errorMessage, err.Error()))
		return
	}

	var response string
	switch findState {
	case parser.StateFoundAndResolved:
		response = fmt.Sprintf("%s - Dosar găsit și rezolvat. Poți continua procedurile ulterioare.", decreeNumber)
	case parser.StateFoundButNotResolved:
		response = fmt.Sprintf("%s - Dosar găsit dar nerezolvat încă. Va trebui să mai aștepți.", decreeNumber)
	case parser.StateNotFound:
		response = fmt.Sprintf("%s - Dosar negăsit. Te rog verifică numărul și anul.", decreeNumber)
	default:
		response = unknownState
	}

	h.sendMessage(ctx, b, update.Message.Chat.ID, response)
}

func (h *BotHandler) sendHelpMessage(ctx context.Context, b *bot.Bot, chatID int64) {
	h.sendMessage(ctx, b, chatID, startMessage)
}

func (h *BotHandler) sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}

// Stop stops the bot
func (h *BotHandler) Stop() {
	if h.botInstance != nil {
		log.Println("Stopping bot...")
		// The go-telegram-bot library doesn't have a direct Stop method,
		// but we can cancel the context in the Init function
	}
}
