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

type TelegramBot interface {
	Init(onMessage func(ctx context.Context, update *models.Update), ctx context.Context) error
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendMessageWithSubscribe(ctx context.Context, chatID int64, text, decreeNumber string) error
}

type botHandler struct {
	instance            *bot.Bot
	subscriptionService database.SubscriptionService
}

func newBotHandler(
	subscriptionService database.SubscriptionService,
) TelegramBot {
	return &botHandler{
		subscriptionService: subscriptionService,
	}
}

func (h *botHandler) Init(onMessage func(ctx context.Context, update *models.Update), ctx context.Context) error {
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
		bot.WithMessageTextHandler("/my", bot.MatchTypeExact, h.listSubscriptionsCommand),
		bot.WithMessageTextHandler("/add_subscribe", bot.MatchTypeExact, h.addSubscriptionCommand),
		bot.WithMessageTextHandler("/remove_subscribe", bot.MatchTypeExact, h.removeSubscriptionCommand),
		bot.WithMessageTextHandler("/remove_all_subscribe", bot.MatchTypeExact, h.removeAllSubscriptionsCommand),
	}

	var err error
	h.instance, err = bot.New(token, opts...)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	log.Println("🤖 Starting Telegram bot...")

	// h.SendMessage(context.TODO(), 574037714, "Bot started")

	h.instance.Start(ctx)

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

func (h *botHandler) SendMessageWithSubscribe(ctx context.Context, chatID int64, text, decreeNumber string) error {
	kb := inline.New(h.instance).Row().Button("Adauga la notificari", []byte(fmt.Sprintf("/add_subscribe %v", decreeNumber)), h.onInlineKeyboardSelect)

	_, err := h.instance.SendMessage(context.TODO(), &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})

	return err
}

func (h *botHandler) listSubscriptionsCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	subscriptions, err := h.subscriptionService.GetSubscriptions(update.Message.Chat.ID)
	fmt.Println(subscriptions)
	if err != nil {
		h.SendMessage(ctx, update.Message.Chat.ID, "Error fetching subscriptions")
		return
	}
	if len(subscriptions) == 0 {
		h.SendMessage(ctx, update.Message.Chat.ID, "No subscriptions found")
		return
	}
	var response string
	for _, subscription := range subscriptions {
		response += fmt.Sprintf("Decree number: %s\n", subscription)
	}
	h.SendMessage(ctx, update.Message.Chat.ID, response)
}
func (h *botHandler) addSubscriptionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.subscriptionService.CreateSubscription(update.Message.Chat.ID, "decree_number")
	h.SendMessage(ctx, update.Message.Chat.ID, "Subscription added")
}

func (h *botHandler) removeSubscriptionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.subscriptionService.DeleteSubscription(update.Message.Chat.ID, "decree_number")
	h.SendMessage(ctx, update.Message.Chat.ID, "Subscription removed")
}

func (h *botHandler) removeAllSubscriptionsCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.subscriptionService.DeleteAllSubscriptions(update.Message.Chat.ID)
	h.SendMessage(ctx, update.Message.Chat.ID, "All subscriptions removed")
}

func (h *botHandler) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	decree_number := strings.Trim(strings.Split(string(data), "/add_subscribe")[1], " ")

	h.subscriptionService.CreateSubscription(mes.Message.Chat.ID, decree_number)
	h.SendMessage(ctx, mes.Message.Chat.ID, fmt.Sprintf("Subscription added for decree number: %s", decree_number))
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
