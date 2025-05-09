package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andiq123/cetatenie-analyzer/internal/bot"
	"github.com/joho/godotenv"
)

func init() {
	// Try loading .env file but don't fail if it's missing
	if err := godotenv.Load(); err != nil {
		log.Printf("Notice: Could not load .env file (%v) - using environment variables", err)
	}
}

func main() {
	// Initialize the bot with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the bot in a goroutine
	go func() {
		bot.Init(ctx)
	}()

	// Wait for shutdown signal
	waitForShutdown(cancel)
}

func waitForShutdown(cancel context.CancelFunc) {
	// Set up channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-sigs
	log.Printf("Received signal %v, shutting down gracefully...", sig)

	// Cancel the main context
	cancel()

	log.Println("Shutdown complete")
	os.Exit(0)
}
