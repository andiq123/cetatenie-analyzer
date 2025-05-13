package database

type Subscription struct {
	ID           uint `gorm:"primaryKey"`
	ChatID       int64
	DecreeNumber string `gorm:"uniqueIndex"`
}
