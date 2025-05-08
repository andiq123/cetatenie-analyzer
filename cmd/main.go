package main

import (
	"log"

	"github.com/andiq123/cetatenie-analyzer/internal/bot"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
}

func main() {
	log.Println("Starting Cetățenie Analyzer Bot...")
	bot.Init()
}
