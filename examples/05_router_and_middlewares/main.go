// Package main provides an example.
package main

import (
	"context"
	"log"
	"time"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func main() {
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", yabotapi.WithDebug(true))
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	// Initialize FSM (Finite State Machine) storage with TTL
	storage := yabotapi.NewMemoryStorage(10 * time.Minute)
	defer storage.Stop()

	// Initialize the router and attach the FSM storage
	r := router.NewRouter(bot).WithStorage(storage)

	// Example of using a Middleware (global interceptor)
	r.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			u := c.Update
			log.Printf("[Middleware] Incoming update ID: %d", u.UpdateID)

			// Pass control to the next handler
			err := next(c)

			log.Printf("[Middleware] Finished processing ID: %d", u.UpdateID)
			return err
		}
	})

	r.HandleCommand("start", func(c *router.Context) error {
		return c.Reply("Hello! Send /ask to start the questionnaire.")
	})

	r.HandleCommand("ask", func(c *router.Context) error {
		// Set the state for the user
		c.FSM().SetState("waiting_for_name")
		return c.Reply("What is your name?")
	})

	// Handler that catches updates ONLY when the user is in the "waiting_for_name" state
	r.HandleState("waiting_for_name", func(c *router.Context) error {
		name := c.Update.Text

		// Clear the state (revert the user back to the default state)
		c.FSM().Clear()
		return c.Reply("Nice to meet you, " + name + "!")
	})

	// Fallback handler for all other cases
	r.HandleDefault(func(_ *router.Context) error {
		log.Println("Received an update that does not have a specific handler.")
		return nil
	})

	log.Println("Bot is running. Send '/ask' to test the router, FSM, and Middleware.")
	if err := r.StartPolling(context.Background(), 0); err != nil {
		log.Println("Polling stopped with error:", err)
	}
}
