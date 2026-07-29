// Package main provides an example.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
)

func main() {
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", yabotapi.WithDebug(true))
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	ctx := context.Background()

	// The URL of your server, which must be publicly accessible (HTTPS)
	webhookURL := "https://your-domain.com/webhook"

	// Set the webhook in Yandex
	if err := bot.Webhooks.SetWebhook(ctx, webhookURL); err != nil {
		log.Fatal("Failed to set webhook:", err)
	}
	log.Printf("Webhook successfully set to: %s", webhookURL)

	// Register the handler and get the updates channel
	updatesChannel, handler := bot.Webhooks.ListenForWebhook(ctx)
	http.HandleFunc("/webhook", handler)

	// Start the HTTP server in a separate goroutine
	go func() {
		log.Println("Server is listening on port 8080...")
		srv := &http.Server{Addr: ":8080", ReadHeaderTimeout: 3 * time.Second}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("HTTP server error:", err)
		}
	}()

	// Process updates from the channel
	for update := range updatesChannel {
		if update.Text != "" {
			reply := yabotapi.NewReply(&update, "Received via Webhook: "+update.Text)
			if _, err := bot.Messages.SendText(ctx, reply); err != nil {
				log.Println("Error sending message:", err)
			}
		}
	}
}
