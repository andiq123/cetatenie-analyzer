package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andiq123/cetatenie-analyzer/internal/bot"
	"github.com/andiq123/cetatenie-analyzer/internal/decree_processor"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	dp := decree_processor.New()
	handler := bot.New(dp)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		if err := handler.Close(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	if err := handler.Init(); err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}
}
