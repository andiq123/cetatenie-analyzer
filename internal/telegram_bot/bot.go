package telegram_bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

// Command constants - fixed to comply with Telegram's requirements
// Commands must be all lowercase English letters, digits, and underscores
const (
	cmdStart                  = "/start"
	cmdHelp                   = "/ajutor"
	cmdMySubscriptions        = "/abonamente"
	cmdAddSubscription        = "/adauga"
	cmdRemoveSubscription     = "/sterge"
	cmdRemoveAllSubscriptions = "/sterge_toate"
)

var botCommands = []models.BotCommand{
	{Command: strings.TrimPrefix(cmdStart, "/"), Description: "Pornire bot și mesaj de bun venit"},
	{Command: strings.TrimPrefix(cmdHelp, "/"), Description: "Ajutor și informații despre comenzi"},
	{Command: strings.TrimPrefix(cmdMySubscriptions, "/"), Description: "Listează toate abonamentele tale"},
	{Command: strings.TrimPrefix(cmdAddSubscription, "/"), Description: "Adaugă un abonament la un dosar"},
	{Command: strings.TrimPrefix(cmdRemoveSubscription, "/"), Description: "Șterge un abonament la un dosar"},
	{Command: strings.TrimPrefix(cmdRemoveAllSubscriptions, "/"), Description: "Șterge toate abonamentele"},
}

// TelegramBot defines the interface for the Telegram bot functionality
type TelegramBot interface {
	Init(onMessage func(ctx context.Context, update *models.Update), ctx context.Context) error
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendMessageWithSubscribe(ctx context.Context, chatID int64, text, decreeNumber string) error
}

// botHandler implements the TelegramBot interface
type botHandler struct {
	instance            *bot.Bot
	subscriptionService database.SubscriptionService
}

// NewBotHandler creates a new instance of the Telegram bot handler
func NewBotHandler(subscriptionService database.SubscriptionService) TelegramBot {
	return &botHandler{
		subscriptionService: subscriptionService,
	}
}

// Init initializes the bot with the provided token and sets up command handlers
func (h *botHandler) Init(onMessage func(ctx context.Context, update *models.Update), ctx context.Context) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("variabila de mediu TELEGRAM_BOT_TOKEN nu este setată")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			onMessage(ctx, update)
		}),
		bot.WithMessageTextHandler(cmdStart, bot.MatchTypeExact, h.startCommand),
		bot.WithMessageTextHandler(cmdHelp, bot.MatchTypeExact, h.helpCommand),
		bot.WithMessageTextHandler(cmdMySubscriptions, bot.MatchTypeExact, h.listSubscriptionsCommand),
		bot.WithMessageTextHandler(cmdAddSubscription, bot.MatchTypePrefix, h.addSubscriptionCommand),
		bot.WithMessageTextHandler(cmdRemoveSubscription, bot.MatchTypePrefix, h.removeSubscriptionCommand),
		bot.WithMessageTextHandler(cmdRemoveAllSubscriptions, bot.MatchTypeExact, h.removeAllSubscriptionsCommand),
	}

	var err error
	h.instance, err = bot.New(token, opts...)
	if err != nil {
		return fmt.Errorf("eroare la crearea botului: %w", err)
	}

	// Set the bot commands with all parameters
	_, err = h.instance.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands:     botCommands,
		Scope:        &models.BotCommandScopeDefault{}, // Default scope for all chats
		LanguageCode: "ro",                             // Romanian language code
	})
	if err != nil {
		return fmt.Errorf("eroare la setarea comenzilor botului: %w", err)
	}

	log.Println("🤖 Pornire bot Telegram...")
	h.instance.Start(ctx)

	return nil
}

// SendMessage sends a message to a specific chat
func (h *botHandler) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := h.instance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	return err
}

// SendMessageWithSubscribe sends a message with a subscription button
func (h *botHandler) SendMessageWithSubscribe(ctx context.Context, chatID int64, text, decreeNumber string) error {
	kb := inline.New(h.instance).Row().Button("Adaugă la notificări", []byte(fmt.Sprintf("%s %s", cmdAddSubscription, decreeNumber)), h.onInlineKeyboardSelect)

	_, err := h.instance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
	return err
}

// Command handlers
func (h *botHandler) listSubscriptionsCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	subscriptions, err := h.subscriptionService.GetSubscriptions(update.Message.Chat.ID)
	if err != nil {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ Eroare la obținerea abonamentelor")
		return
	}
	if len(subscriptions) == 0 {
		h.SendMessage(ctx, update.Message.Chat.ID, "📭 Nu ai niciun abonament activ")
		return
	}

	var response strings.Builder
	response.WriteString("📋 *Abonamentele tale:*\n\n")
	for _, subscription := range subscriptions {
		response.WriteString(fmt.Sprintf("• Dosar: `%s`\n", subscription))
	}
	h.SendMessage(ctx, update.Message.Chat.ID, response.String())
}

func (h *botHandler) addSubscriptionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := strings.Fields(update.Message.Text)
	if len(args) < 2 {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ Te rog specifică numărul dosarului")
		return
	}

	decreeNumber := args[1]
	err := h.subscriptionService.CreateSubscription(update.Message.Chat.ID, decreeNumber)
	if err != nil {
		if strings.Contains(err.Error(), "subscription already exists") {
			h.SendMessage(ctx, update.Message.Chat.ID, fmt.Sprintf("ℹ️ Ești deja abonat la dosarul `%s`", decreeNumber))
			return
		}
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ Eroare la adăugarea abonamentului")
		return
	}

	h.SendMessage(ctx, update.Message.Chat.ID, fmt.Sprintf("✅ Abonament adăugat pentru dosarul `%s`", decreeNumber))
}

func (h *botHandler) removeSubscriptionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := strings.Fields(update.Message.Text)
	if len(args) < 2 {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ Te rog specifică numărul dosarului")
		return
	}

	decreeNumber := args[1]
	if err := h.subscriptionService.DeleteSubscription(update.Message.Chat.ID, decreeNumber); err != nil {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ Eroare la ștergerea abonamentului")
		return
	}

	h.SendMessage(ctx, update.Message.Chat.ID, fmt.Sprintf("✅ Abonament șters pentru dosarul `%s`", decreeNumber))
}

func (h *botHandler) removeAllSubscriptionsCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if err := h.subscriptionService.DeleteAllSubscriptions(update.Message.Chat.ID); err != nil {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ Eroare la ștergerea abonamentelor")
		return
	}

	h.SendMessage(ctx, update.Message.Chat.ID, "✅ Toate abonamentele au fost șterse")
}

func (h *botHandler) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	decreeNumber := strings.Trim(strings.Split(string(data), cmdAddSubscription)[1], " ")

	err := h.subscriptionService.CreateSubscription(mes.Message.Chat.ID, decreeNumber)
	if err != nil {
		if strings.Contains(err.Error(), "subscription already exists") {
			h.SendMessage(ctx, mes.Message.Chat.ID, fmt.Sprintf("ℹ️ Ești deja abonat la dosarul `%s`", decreeNumber))
			return
		}
		h.SendMessage(ctx, mes.Message.Chat.ID, "❌ Eroare la adăugarea abonamentului")
		return
	}

	h.SendMessage(ctx, mes.Message.Chat.ID, fmt.Sprintf("✅ Abonament adăugat pentru dosarul `%s`", decreeNumber))
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
