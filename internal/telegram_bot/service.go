package telegram_bot

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/andiq123/cetatenie-analyzer/internal/decree"
	"github.com/andiq123/cetatenie-analyzer/internal/timer"
	"github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

const (
	errorSendingMessage   = "error sending message: %w"
	errorProcessingDecree = "error processing decree: %w"
)

type BotService interface {
	Start(ctx context.Context) error
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type botService struct {
	bh        TelegramBot
	processor decree.Processor
}

func NewBot(db *gorm.DB) BotService {
	return &botService{
		processor: decree.NewProcessor(),
		bh:        NewBotHandler(database.NewSubscriptionService(db)),
	}
}

func (b *botService) Start(ctx context.Context) error {
	if err := b.bh.Init(b.defaultHandler, ctx); err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}
	return nil
}

func (b *botService) SendMessage(ctx context.Context, chatID int64, text string) error {
	if err := b.bh.SendMessage(ctx, chatID, text); err != nil {
		return fmt.Errorf(errorSendingMessage, err)
	}
	return nil
}

func (b *botService) defaultHandler(ctx context.Context, update *models.Update) {
	if !regexp.MustCompile(decreePattern).MatchString(update.Message.Text) {
		if err := b.bh.SendMessage(ctx, update.Message.Chat.ID, invalidFormat); err != nil {
			fmt.Printf("Error sending invalid format message: %v\n", err)
		}
		return
	}

	b.handleDecreeRequest(ctx, update)
}

func (b *botService) handleDecreeRequest(ctx context.Context, update *models.Update) {
	senderId := update.Message.Chat.ID
	decreeNumber := strings.TrimSpace(update.Message.Text)

	if err := b.bh.SendMessage(ctx, senderId, fmt.Sprintf(searching, decreeNumber)); err != nil {
		fmt.Printf("Error sending searching message: %v\n", err)
		return
	}

	findState, timeReport, err := b.processor.Handle(decreeNumber)
	if err != nil {
		if err := b.bh.SendMessage(ctx, senderId, fmt.Sprintf(errorMessage, err.Error())); err != nil {
			fmt.Printf("Error sending error message: %v\n", err)
		}
		return
	}

	var response string
	switch findState {
	case decree.StateFoundAndResolved:
		response = fmt.Sprintf(successMessage, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	case decree.StateFoundButNotResolved:
		response = fmt.Sprintf(inProgressMsg, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
		if err := b.bh.SendMessageWithSubscribe(ctx, senderId, response, decreeNumber); err != nil {
			fmt.Printf("Error sending message with subscribe: %v\n", err)
			return
		}
		return
	case decree.StateNotFound:
		response = fmt.Sprintf(notFoundMsg, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	default:
		response = unknownState
	}

	if err := b.bh.SendMessage(ctx, senderId, response); err != nil {
		fmt.Printf("Error sending response message: %v\n", err)
	}
}
