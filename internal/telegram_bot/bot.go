package telegram_bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

// Command constants - fixed to comply with Telegram's requirements
// Commands must be all lowercase English letters, digits, and underscores
const (
	cmdStart                  = "start"
	cmdHelp                   = "ajutor"
	cmdMySubscriptions        = "abonamente"
	cmdAddSubscription        = "adauga"
	cmdRemoveSubscription     = "sterge"
	cmdRemoveAllSubscriptions = "sterge_toate"
)

var botCommands = []models.BotCommand{
	{Command: cmdStart, Description: "🎯 Pornește botul și vezi mesajul de bun venit"},
	{Command: cmdHelp, Description: "❓ Vezi ajutor și informații despre comenzi"},
	{Command: cmdMySubscriptions, Description: "📋 Vezi toate dosarele la care ești abonat"},
	{Command: cmdAddSubscription, Description: "➕ Adaugă un dosar la notificări (ex: /adauga 123/RD/2023)"},
	{Command: cmdRemoveSubscription, Description: "➖ Șterge un dosar din notificări (ex: /sterge 123/RD/2023)"},
	{Command: cmdRemoveAllSubscriptions, Description: "🗑 Șterge toate abonamentele la dosare"},
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
		bot.WithMessageTextHandler("/"+cmdStart, bot.MatchTypeExact, h.startCommand),
		bot.WithMessageTextHandler("/"+cmdHelp, bot.MatchTypeExact, h.helpCommand),
		bot.WithMessageTextHandler("/"+cmdMySubscriptions, bot.MatchTypeExact, h.listSubscriptionsCommand),
		bot.WithMessageTextHandler("/"+cmdRemoveAllSubscriptions, bot.MatchTypeExact, h.removeAllSubscriptionsCommand),
		bot.WithMessageTextHandler("/"+cmdAddSubscription, bot.MatchTypePrefix, h.addSubscriptionCommand),
		bot.WithMessageTextHandler("/"+cmdRemoveSubscription, bot.MatchTypePrefix, h.removeSubscriptionCommand),
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
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Eroare la obținerea abonamentelor</b>\n\nTe rugăm să încerci din nou mai târziu.")
		return
	}
	if len(subscriptions) == 0 {
		h.SendMessage(ctx, update.Message.Chat.ID, "📭 <b>Nu ai niciun abonament activ</b>\n\nFolosește comanda /adauga pentru a adăuga un dosar la notificări.")
		return
	}

	var response strings.Builder
	response.WriteString("📋 <b>Abonamentele tale:</b>\n\n")
	for _, subscription := range subscriptions {
		response.WriteString(fmt.Sprintf("• Dosar: <code>%s</code>\n", subscription))
	}
	h.SendMessage(ctx, update.Message.Chat.ID, response.String())
}

func (h *botHandler) addSubscriptionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Split the message into command and arguments
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Format invalid</b>\n\nTe rog specifică numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\nExemplu: <code>123/RD/2023</code>")
		return
	}

	// Get the decree number (everything after the command)
	decreeNumber := strings.Join(parts[1:], " ")

	// Validate the decree number format
	if !regexp.MustCompile(decreePattern).MatchString(decreeNumber) {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Format invalid</b>\n\nTe rog specifică numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\nExemplu: <code>123/RD/2023</code>")
		return
	}

	err := h.subscriptionService.CreateSubscription(update.Message.Chat.ID, decreeNumber)
	if err != nil {
		if strings.Contains(err.Error(), "subscription already exists") {
			h.SendMessage(ctx, update.Message.Chat.ID, fmt.Sprintf("ℹ️ <b>Abonament existent</b>\n\nEști deja abonat la dosarul <code>%s</code>", decreeNumber))
			return
		}
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Eroare la adăugarea abonamentului</b>\n\nTe rugăm să încerci din nou mai târziu.")
		return
	}

	h.SendMessage(ctx, update.Message.Chat.ID, fmt.Sprintf("✅ <b>Abonament adăugat</b>\n\nAi fost abonat cu succes la dosarul <code>%s</code>", decreeNumber))
}

func (h *botHandler) removeSubscriptionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Split the message into command and arguments
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Format invalid</b>\n\nTe rog specifică numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\nExemplu: <code>123/RD/2023</code>")
		return
	}

	// Get the decree number (everything after the command)
	decreeNumber := strings.Join(parts[1:], " ")

	// Validate the decree number format
	if !regexp.MustCompile(decreePattern).MatchString(decreeNumber) {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Format invalid</b>\n\nTe rog specifică numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\nExemplu: <code>123/RD/2023</code>")
		return
	}

	if err := h.subscriptionService.DeleteSubscription(update.Message.Chat.ID, decreeNumber); err != nil {
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Eroare la ștergerea abonamentului</b>\n\nTe rugăm să încerci din nou mai târziu.")
		return
	}

	h.SendMessage(ctx, update.Message.Chat.ID, fmt.Sprintf("✅ <b>Abonament șters</b>\n\nAi fost dezabonat cu succes de la dosarul <code>%s</code>", decreeNumber))
}

func (h *botHandler) removeAllSubscriptionsCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Add logging to debug the command
	log.Printf("Received removeAllSubscriptionsCommand from chat ID: %d", update.Message.Chat.ID)

	if err := h.subscriptionService.DeleteAllSubscriptions(update.Message.Chat.ID); err != nil {
		log.Printf("Error deleting all subscriptions: %v", err)
		h.SendMessage(ctx, update.Message.Chat.ID, "❌ <b>Eroare la ștergerea abonamentelor</b>\n\nTe rugăm să încerci din nou mai târziu.")
		return
	}

	log.Printf("Successfully deleted all subscriptions for chat ID: %d", update.Message.Chat.ID)
	h.SendMessage(ctx, update.Message.Chat.ID, "✅ <b>Abonamente șterse</b>\n\nToate abonamentele tale au fost șterse cu succes.")
}

func (h *botHandler) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	// Extract the decree number from the data
	parts := strings.Split(string(data), " ")
	if len(parts) < 2 {
		h.SendMessage(ctx, mes.Message.Chat.ID, "❌ <b>Eroare la procesarea comenzii</b>\n\nTe rugăm să încerci din nou.")
		return
	}

	decreeNumber := parts[1]

	// Validate the decree number format
	if !regexp.MustCompile(decreePattern).MatchString(decreeNumber) {
		h.SendMessage(ctx, mes.Message.Chat.ID, "❌ <b>Format invalid</b>\n\nTe rog specifică numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\nExemplu: <code>123/RD/2023</code>")
		return
	}

	err := h.subscriptionService.CreateSubscription(mes.Message.Chat.ID, decreeNumber)
	if err != nil {
		if strings.Contains(err.Error(), "subscription already exists") {
			h.SendMessage(ctx, mes.Message.Chat.ID, fmt.Sprintf("ℹ️ <b>Abonament existent</b>\n\nEști deja abonat la dosarul <code>%s</code>", decreeNumber))
			return
		}
		h.SendMessage(ctx, mes.Message.Chat.ID, "❌ <b>Eroare la adăugarea abonamentului</b>\n\nTe rugăm să încerci din nou mai târziu.")
		return
	}

	h.SendMessage(ctx, mes.Message.Chat.ID, fmt.Sprintf("✅ <b>Abonament adăugat</b>\n\nAi fost abonat cu succes la dosarul <code>%s</code>", decreeNumber))
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
