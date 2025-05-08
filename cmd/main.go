package main

import (
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

	// Start the health check server
	startHealthCheckServer()

	// Initialize the bot
	bot.Init()

	// Keep the main goroutine alive
	waitForShutdown()
}

func startHealthCheckServer() {
	// Create a simple HTTP server for health checks
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Also handle the specific path from the logs
	http.HandleFunc("/kaithheathcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Get the port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Starting health check server on port %s...", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Health check server failed: %v", err)
		}
	}()
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
