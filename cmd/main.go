package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/andiq123/cetatenie-analyzer/internal/database"
	"github.com/andiq123/cetatenie-analyzer/internal/telegram_bot"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	db, err := database.InitDb()
	if err != nil {
		panic(err)
	}

	bot := telegram_bot.NewBot(db)
	err = bot.Start(ctx)
	if err != nil {
		panic(err)
	}

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// go func() {
	// 	<-sigChan
	// 	log.Println("Shutting down...")
	// 	if err := handler.Close(); err != nil {
	// 		log.Printf("Error during shutdown: %v", err)
	// 	}
	// 	os.Exit(0)
	// }()

	// if err := handler.Init(); err != nil {
	// 	log.Fatalf("Failed to initialize bot: %v", err)
	// }
}
