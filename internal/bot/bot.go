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
	startMessage  = `🌟 *Bun venit la Cetățenie Analyzer!* 🇷🇴

Cu acest bot poți verifica starea dosarului tău de redobândire a cetățeniei române. 

_Cum funcționează?_ 🤔
1. Trimite numărul dosarului în formatul: *[număr]/RD/[an]*
   Exemplu: ` + "`123/RD/2023`" + `
2. Așteaptă rezultatul
3. Dacă ai nevoie de ajutor, apasă pe butonul *"Meniu"* sau tastează /help

Succes în procesul tău! 🍀`
	invalidFormat = "❌ *Format invalid* \n\nTe rog folosește formatul: `[număr]/RD/[an]`\n\nExemplu: `123/RD/2023`"
	searching     = "🔍 _Caut dosarul:_ `%s`\n\nTe rog așteaptă puțin..."
	errorMessage  = "⚠️ *A apărut o eroare*: \n\n`%s`\n\nTe rugăm să încerci din nou mai târziu."
	unknownState  = "❓ *Stare necunoscută*\n\nTe rugăm să încerci mai târziu sau să contactezi administratorul."
	helpMessage   = `ℹ️ *Ajutor și instrucțiuni*

📌 _Cum verific dosarul?_
Trimite numărul dosarului în formatul: *[număr]/RD/[an]*
Exemplu: ` + "`123/RD/2023`" + `

📌 _Ce înseamnă rezultatele?_
✅ *Găsit și rezolvat* - Dosar finalizat, poți continua procedurile
🔄 *Găsit dar nerezolvat* - Dosar în procesare, mai așteaptă
❌ *Negăsit* - Verifică numărul sau contactează autoritățile

📌 _Comenzi disponibile:_
/start - Mesaj de bun venit
/help - Acest mesaj de ajutor
/menu - Afișează meniul principal

Alte întrebări? Scrie-ne aici! ✉️`
)

type BotHandler struct {
	decreeProcessor decree_processor.Processor
	botInstance     *bot.Bot
}

// Init initializes and starts the Telegram bot with context support
func Init(ctx context.Context) {
	handler := &BotHandler{
		decreeProcessor: decree_processor.New(),
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler.defaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, handler.startCommand),
		bot.WithMessageTextHandler("/help", bot.MatchTypeExact, handler.helpCommand),
		bot.WithMessageTextHandler("/menu", bot.MatchTypeExact, handler.menuCommand),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	handler.botInstance = b

	log.Println("🤖 Starting Telegram bot...")

	b.Start(ctx)
	log.Println("🛑 Telegram bot stopped")
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
	h.sendMenu(ctx, b, update.Message.Chat.ID)
}

func (h *BotHandler) startCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendWelcomeMessage(ctx, b, update.Message.Chat.ID)
}

func (h *BotHandler) helpCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendMessage(ctx, b, update.Message.Chat.ID, helpMessage)
}

func (h *BotHandler) menuCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendMenu(ctx, b, update.Message.Chat.ID)
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
		response = fmt.Sprintf("🎉 *Felicitări!* \n\nDosarul `%s` a fost *găsit și rezolvat*. \n\nPoți continua cu procedurile ulterioare pentru redobândirea cetățeniei!", decreeNumber)
	case parser.StateFoundButNotResolved:
		response = fmt.Sprintf("⏳ *Dosar în procesare* \n\nDosarul `%s` a fost *găsit dar nu este rezolvat încă*. \n\nVa trebui să mai aștepți până când va fi finalizat.", decreeNumber)
	case parser.StateNotFound:
		response = fmt.Sprintf("🔎 *Rezultat negativ* \n\nDosarul `%s` *nu a fost găsit*. \n\nTe rugăm să verifici numărul și anul, sau să contactezi autoritățile competente.", decreeNumber)
	default:
		response = unknownState
	}

	h.sendMessage(ctx, b, update.Message.Chat.ID, response)
	h.sendMenu(ctx, b, update.Message.Chat.ID)
}

func (h *BotHandler) sendWelcomeMessage(ctx context.Context, b *bot.Bot, chatID int64) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      startMessage,
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "📋 Meniu", CallbackData: "menu"},
					{Text: "ℹ️ Ajutor", CallbackData: "help"},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (h *BotHandler) sendMenu(ctx context.Context, b *bot.Bot, chatID int64) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "📱 *Meniu Principal* - Alege o opțiune:",
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{Text: "🔍 Verifică dosar"},
				},
				{
					{Text: "ℹ️ Ajutor"},
					{Text: "🏠 Acasă"},
				},
			},
			ResizeKeyboard: true,
		},
	})
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (h *BotHandler) sendHelpMessage(ctx context.Context, b *bot.Bot, chatID int64) {
	h.sendMessage(ctx, b, chatID, helpMessage)
}

func (h *BotHandler) sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
