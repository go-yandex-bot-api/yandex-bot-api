// Package main provides an example.
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
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

	// Handle the /start command
	r.HandleCommand("start", func(c *router.Context) error {
		keyboard := yabotapi.NewSuggestButtonsGrid(true,
			[]yabotapi.InlineSuggestButton{
				yabotapi.NewSimpleActionButton("Say Hi", "btn_hi"),
				yabotapi.NewSimpleActionButton("Say Bye", "btn_bye"),
			},
			[]yabotapi.InlineSuggestButton{
				yabotapi.NewURLButton("Open Yandex", "https://ya.ru"),
				yabotapi.NewTextButton("Send Payload", "Payload Sent!"),
			},
		)

		return c.ReplyWithKeyboard("Hello! Choose an action:", keyboard)
	})

	// Handle button clicks
	r.HandleButton("btn_hi", func(c *router.Context) error {
		return c.Reply("Hi there! 👋")
	})

	r.HandleButton("btn_bye", func(c *router.Context) error {
		return c.Reply("Goodbye! Have a nice day. 🚀")
	})

	log.Println("Send '/start' to get the keyboard.")
	r.Start(ctx, updatesChannel)
}
