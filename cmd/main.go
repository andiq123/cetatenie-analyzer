package main

import (
	"context"
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
	// Try loading .env file but don't fail if it's missing
	if err := godotenv.Load(); err != nil {
		log.Printf("Notice: Could not load .env file (%v) - using environment variables", err)
	}
}

func main() {
	log.Println("Starting Cetățenie Analyzer Bot...")

	// Initialize the bot with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot.Init(ctx)

	// Configure HTTP server with proper timeouts
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      healthCheckHandler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting health check server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Verify server is running
	if err := waitForServerReady(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	waitForShutdown(srv, cancel)
}

func healthCheckHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Add any additional health checks here
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"status\":\"healthy\"}"))
	})
	return mux
}

func waitForServerReady(port string) error {
	client := http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://localhost:%s/health", port)

	// Try for up to 30 seconds
	for i := 0; i < 30; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("server did not become ready in time")
}

func waitForShutdown(srv *http.Server, cancel context.CancelFunc) {
	// Set up channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-sigs
	log.Printf("Received signal %v, shutting down gracefully...", sig)

	// Create shutdown context with timeout
	ctx, cancelTimeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelTimeout()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Cancel the main context
	cancel()

	log.Println("Shutdown complete")
	os.Exit(0)
}
