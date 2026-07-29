// Package main provides an example.
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/examples/09_project_structure/handlers"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func main() {
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", yabotapi.WithDebug(true))
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	ctx := context.Background()

	updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal("Failed to start polling:", err)
	}

	r := router.NewRouter(bot)

	// --- 1. Register simple handlers (from start.go) ---
	handlers.RegisterStartHandlers(r)

	// --- 2. Register complex handlers with Dependency Injection (from echo.go) ---
	// Imagine this is your *sql.DB or Redis client
	dbConn := "PostgreSQL_Fake_Connection"

	echoController := handlers.NewEchoController(dbConn)
	echoController.Register(r)

	log.Println("Modular Bot is running. Send /start or /echo [text]")
	r.Start(ctx, updatesChannel)
}
