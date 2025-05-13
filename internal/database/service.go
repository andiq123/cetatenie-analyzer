package database

import (
	"fmt"

	"gorm.io/gorm"
)

type SubscriptionService interface {
	CreateSubscription(chatID int64, decreeNumber string) error
	DeleteSubscription(chatID int64, decreeNumber string) error
	DeleteAllSubscriptions(chatID int64) error
	GetSubscriptions(chatID int64) ([]string, error)
	GetAllSubscriptions() ([]Subscription, error)
}

type subscriptionService struct {
	db *gorm.DB
}

func NewSubscriptionService(db *gorm.DB) SubscriptionService {
	return &subscriptionService{db: db}
}

func (s *subscriptionService) CreateSubscription(chatID int64, decreeNumber string) error {
	// Check if subscription already exists
	var existingSubscription Subscription
	result := s.db.Where("chat_id = ? AND decree_number = ?", chatID, decreeNumber).First(&existingSubscription)
	if result.Error == nil {
		return fmt.Errorf("subscription already exists for decree number %s", decreeNumber)
	}
	if result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("error checking existing subscription: %v", result.Error)
	}

	subscription := Subscription{
		ChatID:       chatID,
		DecreeNumber: decreeNumber,
	}
	return s.db.Create(&subscription).Error
}

func (s *subscriptionService) DeleteSubscription(chatID int64, decreeNumber string) error {
	return s.db.Where("chat_id = ? AND decree_number = ?", chatID, decreeNumber).Delete(&Subscription{}).Error
}

func (s *subscriptionService) DeleteAllSubscriptions(chatID int64) error {
	subscription := Subscription{
		ChatID: chatID,
	}
	return s.db.Where("chat_id = ?", chatID).Delete(&subscription).Error
}

func (s *subscriptionService) GetSubscriptions(chatID int64) ([]string, error) {
	var subscriptions []Subscription
	err := s.db.Where("chat_id = ?", chatID).Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	var decreeNumbers []string
	for _, subscription := range subscriptions {
		decreeNumbers = append(decreeNumbers, subscription.DecreeNumber)
	}
	return decreeNumbers, nil
}

func (s *subscriptionService) GetAllSubscriptions() ([]Subscription, error) {
	var subscriptions []Subscription
	err := s.db.Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}
