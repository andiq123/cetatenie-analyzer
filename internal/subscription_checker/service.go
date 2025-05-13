package subscription_checker

import (
	"context"
	"fmt"
	"time"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/andiq123/cetatenie-analyzer/internal/decree"
	"github.com/andiq123/cetatenie-analyzer/internal/telegram_bot"
)

const (
	operationTimeout          = 30 * time.Second
	errorGettingSubscriptions = "error getting subscriptions: %w"
	errorCheckingDecree       = "error checking decree: %w"
	errorSendingMessage       = "error sending message: %w"
	errorRemovingSubscription = "error removing subscription: %w"
)

// Service defines the interface for subscription checking functionality
type Service interface {
	CheckAllSubscriptions() error
}

// service implements the Service interface
type service struct {
	subscriptionService database.SubscriptionService
	decreeService       decree.Processor
	bot                 telegram_bot.BotService
}

// NewService creates a new instance of the subscription checker service
func NewService(subscriptionService database.SubscriptionService, decreeService decree.Processor, bot telegram_bot.BotService) Service {
	return &service{
		subscriptionService: subscriptionService,
		decreeService:       decreeService,
		bot:                 bot,
	}
}

// CheckAllSubscriptions retrieves all subscriptions and checks their states
func (s *service) CheckAllSubscriptions() error {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	subscriptions, err := s.subscriptionService.GetAllSubscriptions()
	if err != nil {
		return fmt.Errorf(errorGettingSubscriptions, err)
	}

	if len(subscriptions) == 0 {
		return nil
	}

	fmt.Printf("Found %d subscriptions to check\n", len(subscriptions))

	for _, sub := range subscriptions {
		if err := s.processSubscription(ctx, sub); err != nil {
			fmt.Printf("Error processing subscription %s: %v\n", sub.DecreeNumber, err)
		}
	}

	return nil
}

func (s *service) processSubscription(ctx context.Context, sub database.Subscription) error {
	state, _, err := s.decreeService.Handle(sub.DecreeNumber)
	if err != nil {
		return fmt.Errorf(errorCheckingDecree, err)
	}

	switch state {
	case decree.StateNotFound:
		return s.handleNotFoundState(ctx, sub)
	case decree.StateFoundAndResolved:
		return s.handleResolvedState(ctx, sub)
	}

	return nil
}

func (s *service) handleNotFoundState(ctx context.Context, sub database.Subscription) error {
	message := fmt.Sprintf("⚠️ <b>Notificare</b>\n\nDosarul <code>%s</code> <b>nu a fost găsit</b>.\n\nTe rugăm să verifici numărul și anul, sau să contactezi autoritățile competente.", sub.DecreeNumber)
	if err := s.bot.SendMessage(ctx, sub.ChatID, message); err != nil {
		return fmt.Errorf(errorSendingMessage, err)
	}
	fmt.Printf("Successfully sent notification to chat %d for decree %s\n", sub.ChatID, sub.DecreeNumber)
	return nil
}

func (s *service) handleResolvedState(ctx context.Context, sub database.Subscription) error {
	message := fmt.Sprintf("🎉 <b>Notificare</b>\n\nDosarul <code>%s</code> <b>a fost găsit și rezolvat</b>!\n\nAcest abonament va fi șters automat.", sub.DecreeNumber)
	if err := s.bot.SendMessage(ctx, sub.ChatID, message); err != nil {
		return fmt.Errorf(errorSendingMessage, err)
	}
	fmt.Printf("Successfully sent notification to chat %d for decree %s\n", sub.ChatID, sub.DecreeNumber)

	if err := s.subscriptionService.DeleteSubscription(sub.ChatID, sub.DecreeNumber); err != nil {
		return fmt.Errorf(errorRemovingSubscription, err)
	}
	fmt.Printf("Successfully removed subscription for decree %s\n", sub.DecreeNumber)
	return nil
}
