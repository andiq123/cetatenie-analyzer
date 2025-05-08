package main

import (
	"fmt"
	"log"
	"net/http"
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

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "8000" // Default port
		}
		http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	}()

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
