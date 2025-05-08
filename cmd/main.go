package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andiq123/cetatenie-analyzer/internal/bot"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
		// Continue even if .env file is not found - we might have env vars set directly
	}
}

func main() {
	log.Println("Starting Cetățenie Analyzer Bot...")

	// Initialize the bot
	bot.Init()

	// Keep the main goroutine alive
	waitForShutdown()
}

func waitForShutdown() {
	// Set up channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-sigs
	log.Printf("Received signal %v, shutting down...", sig)

	// Allow some time for graceful shutdown
	time.Sleep(2 * time.Second)
	log.Println("Shutdown complete")
}
