// Package main provides an example.
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

// PaginationPayload represents the JSON payload we attach to inline buttons.
type PaginationPayload struct {
	Action string `json:"action"`
	Page   int    `json:"page"`
}

const MaxPages = 5

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

	// Step 1: User types /start -> send the first page
	r.HandleCommand("start", func(c *router.Context) error {
		return sendPage(c, 1)
	})

	// Step 2: User clicks a pagination button
	// Note: In Yandex, the action name must match the ServerAction directive's Name
	r.HandleButton("paginate", func(c *router.Context) error {
		var payload PaginationPayload

		// BindPayload automatically unmarshals the JSON attached to the button
		if err := c.BindPayload(&payload); err != nil {
			log.Println("Failed to parse payload:", err)
			return err
		}

		// Send the requested page
		return sendPage(c, payload.Page)
	})

	log.Println("Pagination Bot is running. Send /start")
	r.Start(ctx, updatesChannel)
}

// sendPage builds the keyboard for the specific page and sends it.
func sendPage(c *router.Context, page int) error {
	var buttons []yabotapi.InlineSuggestButton

	// Add "Prev" button if we are not on the first page
	if page > 1 {
		btn := yabotapi.NewActionButton("⬅️ Prev", "paginate", PaginationPayload{Action: "paginate", Page: page - 1})
		buttons = append(buttons, btn)
	}

	// Add "Next" button if we are not on the last page
	if page < MaxPages {
		btn := yabotapi.NewActionButton("Next ➡️", "paginate", PaginationPayload{Action: "paginate", Page: page + 1})
		buttons = append(buttons, btn)
	}

	// Create a persistent keyboard (so it doesn't disappear when clicked)
	keyboard := yabotapi.NewSuggestButtonsGrid(true, buttons)

	// Send the message with the keyboard
	return c.ReplyWithKeyboardf(keyboard, "📄 You are viewing Page %d of %d", page, MaxPages)
}
