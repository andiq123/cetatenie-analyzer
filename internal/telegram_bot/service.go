package telegram_bot

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/andiq123/cetatenie-analyzer/internal/decree"
	"github.com/andiq123/cetatenie-analyzer/internal/timer"
	"github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

type BotService interface {
	Start(ctx context.Context) error
}

type botService struct {
	bh        TelegramBot
	processor decree.Processor
}

func NewBot(db *gorm.DB) BotService {
	return &botService{
		processor: decree.NewProcessor(),
		bh:        newBotHandler(database.NewSubscriptionService(db)),
	}
}

func (b *botService) Start(ctx context.Context) error {
	err := b.bh.Init(b.defaultHandler, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (b *botService) defaultHandler(ctx context.Context, update *models.Update) {
	if !regexp.MustCompile(decreePattern).MatchString(update.Message.Text) {
		b.bh.SendMessage(ctx, update.Message.Chat.ID, invalidFormat)
		return
	}

	b.handleDecreeRequest(ctx, update)
}

func (b *botService) handleDecreeRequest(ctx context.Context, update *models.Update) {
	senderId := update.Message.Chat.ID
	decreeNumber := strings.TrimSpace(update.Message.Text)

	if err := b.bh.SendMessage(ctx, senderId, fmt.Sprintf(searching, decreeNumber)); err != nil {
		log.Printf("Error sending searching message: %v", err)
		return
	}

	findState, timeReport, err := b.processor.Handle(decreeNumber)
	if err != nil {
		b.bh.SendMessage(ctx, senderId, fmt.Sprintf(errorMessage, err.Error()))
		return
	}

	var response string
	switch findState {
	case decree.StateFoundAndResolved:
		response = fmt.Sprintf(successMessage, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	case decree.StateFoundButNotResolved:
		response = fmt.Sprintf(inProgressMsg, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
		b.bh.SendMessageWithSubscribe(ctx, senderId, response, decreeNumber)
		return
	case decree.StateNotFound:
		response = fmt.Sprintf(notFoundMsg, decreeNumber, timer.FormatDuration(timeReport.FetchTime), timer.FormatDuration(timeReport.ParseTime))
	default:
		response = unknownState
	}

	if err := b.bh.SendMessage(ctx, senderId, response); err != nil {
		log.Printf("Error sending response message: %v", err)
	}
}
