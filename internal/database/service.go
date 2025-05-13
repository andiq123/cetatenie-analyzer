package database

import "gorm.io/gorm"

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
	subscription := Subscription{
		ChatID:       chatID,
		DecreeNumber: decreeNumber,
	}
	return s.db.Create(&subscription).Error
}

func (s *subscriptionService) DeleteSubscription(chatID int64, decreeNumber string) error {
	subscription := Subscription{
		ChatID:       chatID,
		DecreeNumber: decreeNumber,
	}
	return s.db.Delete(&subscription).Error
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
