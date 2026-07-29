// Package main provides an example.
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
)

func main() {
	// Initialize the bot with debug mode enabled
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", yabotapi.WithDebug(true))
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	ctx := context.Background()

	// Start Long-Polling
	updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal("Failed to start polling:", err)
	}

	log.Println("Bot is running in Long-Polling mode... Press Ctrl+C to stop.")

	// Process incoming updates in a loop
	for update := range updatesChannel {
		log.Printf("Received update with ID: %d", update.UpdateID)

		// If it's a text message (not a system message or a button callback)
		if update.Text != "" {
			// Reply to the user with the same text (echo bot)
			reply := yabotapi.NewReply(&update, "You said: "+update.Text)

			if _, err := bot.Messages.SendText(ctx, reply); err != nil {
				log.Println("Error sending reply:", err)
			}
		}
	}
}
