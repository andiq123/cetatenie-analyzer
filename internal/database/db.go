package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func initDb() error {
	DB, err := gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	DB.AutoMigrate(&Product{})
	return nil
}
