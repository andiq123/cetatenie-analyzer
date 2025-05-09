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
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	decreePattern = `^\d{1,5}/RD/\d{4}$`
	startMessage  = `🌟 *Bun venit la Cetățenie Analyzer\!* 🇷🇴

Cu acest bot poți verifica starea dosarului tău de redobândire a cetățeniei române\. 

_Cum funcționează?_ 🤔
1\. Trimite numărul dosarului în formatul\: *\[număr\]/RD/\[an\]*
   Exemplu\: ` + "`123/RD/2023`" + `
2\. Așteaptă rezultatul
3\. Dacă ai nevoie de ajutor, apasă pe butonul *\"Meniu\"* sau tastează /help

Succes în procesul tău\! 🍀`

	invalidFormat  = "❌ *Format invalid* \n\nTe rog folosește formatul\\: `\\[număr\\]/RD/\\[an\\]`\n\nExemplu\\: `123/RD/2023`"
	searching      = "🔍 _Caut dosarul\\:_ `%s`\n\nTe rog așteaptă puțin\\."
	errorMessage   = "⚠️ *A apărut o eroare\\:* \n\n`%s`\n\nTe rugăm să încerci din nou mai târziu\\."
	unknownState   = "❓ *Stare necunoscută*\n\nTe rugăm să încerci mai târziu sau să contactezi administratorul\\."
	successMessage = "🎉 *Felicitări\\!* \n\nDosarul `%s` a fost *găsit și rezolvat*\\.\n\nPoți continua cu procedurile ulterioare pentru redobândirea cetățeniei române\\."
	inProgressMsg  = "⏳ *Dosar în procesare* \n\nDosarul `%s` a fost *găsit dar nu este rezolvat încă*\\.\n\nVa trebui să mai aștepți până când va fi finalizat\\."
	notFoundMsg    = "🔎 *Rezultat negativ* \n\nDosarul `%s` *nu a fost găsit*\\.\n\nTe rugăm să verifici numărul și anul\\, sau să contactezi autoritățile competente\\."

	helpMessage = `ℹ️ *Ajutor și instrucțiuni*

📌 _Cum verific dosarul?_
Trimite numărul dosarului în formatul\: *\[număr\]/RD/\[an\]*
Exemplu\: ` + "`123/RD/2023`" + `

📌 _Ce înseamnă rezultatele?_
✅ *Găsit și rezolvat* \- Dosar finalizat, poți continua procedurile
🔄 *Găsit dar nerezolvat* \- Dosar în procesare, mai așteaptă
❌ *Negăsit* \- Verifică numărul sau contactează autoritățile

📌 _Comenzi disponibile\:_
/start \- Mesaj de bun venit
/help \- Acest mesaj de ajutor`
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

	findState, err := h.decreeProcessor.Handle(decreeNumber)
	if err != nil {
		h.sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf(errorMessage, err.Error()))
		return
	}

	var response string
	switch findState {
	case parser.StateFoundAndResolved:
		response = fmt.Sprintf(successMessage, decreeNumber)
	case parser.StateFoundButNotResolved:
		response = fmt.Sprintf(inProgressMsg, decreeNumber)
	case parser.StateNotFound:
		response = fmt.Sprintf(notFoundMsg, decreeNumber)
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
