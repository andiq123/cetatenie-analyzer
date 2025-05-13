package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/andiq123/cetatenie-analyzer/internal/decree"
	"github.com/andiq123/cetatenie-analyzer/internal/subscription_checker"
	"github.com/andiq123/cetatenie-analyzer/internal/telegram_bot"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	fmt.Println("Starting application...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.InitDb()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	subscriptionService := database.NewSubscriptionService(db)
	decreeService := decree.NewProcessor()
	bot := telegram_bot.NewBot(db)

	checker := subscription_checker.NewService(subscriptionService, decreeService, bot)

	fmt.Println("Starting subscription checker...")
	checkerErr := make(chan error, 1)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		if err := checker.CheckAllSubscriptions(); err != nil {
			fmt.Printf("Error in initial subscription check: %v\n", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := checker.CheckAllSubscriptions(); err != nil {
					fmt.Printf("Error checking subscriptions: %v\n", err)
				}
			}
		}
	}()

	fmt.Println("Starting Telegram bot...")
	botErr := make(chan error, 1)
	go func() {
		if err := bot.Start(ctx); err != nil {
			botErr <- fmt.Errorf("failed to start bot: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-checkerErr:
		fmt.Printf("Subscription checker error: %v\n", err)
		cancel()
	case err := <-botErr:
		fmt.Printf("Bot error: %v\n", err)
		cancel()
	case sig := <-sigChan:
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
	}

	time.Sleep(time.Second)
}
