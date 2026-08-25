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

		// Edit the current message seamlessly with the new page content & buttons
		return editPage(c, payload.Page)
	})

	log.Println("Pagination Bot is running. Send /start")
	r.Start(ctx, updatesChannel)
}

// buildKeyboard builds the navigation buttons for a given page.
func buildKeyboard(page int) *yabotapi.SuggestButtons {
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

	return yabotapi.NewSuggestButtonsGrid(true, buttons)
}

// sendPage builds the keyboard for the specific page and sends a new message.
func sendPage(c *router.Context, page int) error {
	keyboard := buildKeyboard(page)
	return c.ReplyWithKeyboardf(keyboard, "📄 You are viewing Page %d of %d", page, MaxPages)
}

// editPage edits the current message in-place with the requested page.
func editPage(c *router.Context, page int) error {
	keyboard := buildKeyboard(page)
	return c.EditCurrentMessageWithKeyboardf(keyboard, "📄 You are viewing Page %d of %d (updated in-place)", page, MaxPages)
}
