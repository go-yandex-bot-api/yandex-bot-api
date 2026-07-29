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
	ctx := context.Background()

	updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal("Failed to start polling:", err)
	}

	// 1. Initialize FSM Memory Storage (keeps data in memory for 30 minutes)
	storage := yabotapi.NewMemoryStorage(30 * time.Minute)
	defer storage.Stop()

	// 2. Attach storage to the router
	r := router.NewRouter(bot).WithStorage(storage)

	// Step 1: User types /start -> ask for Name
	r.HandleCommand("start", func(c *router.Context) error {
		c.FSM().SetState("step_name")
		return c.Reply("Hello! Let's get to know each other. What is your name?")
	})

	// Step 2: Handle Name -> ask for Age
	r.HandleState("step_name", func(c *router.Context) error {
		name := c.Update.Text

		// Save the name into the FSM payload
		c.FSM().SetData("name", name)

		// Move to the next state
		c.FSM().SetState("step_age")
		return c.Replyf("Nice to meet you, %s! How old are you?", name)
	})

	// Step 3: Handle Age -> ask for Favorite Color
	r.HandleState("step_age", func(c *router.Context) error {
		age := c.Update.Text

		c.FSM().SetData("age", age)
		c.FSM().SetState("step_color")
		return c.Reply("Got it. And what is your favorite color?")
	})

	// Step 4: Handle Color -> print summary and clear FSM
	r.HandleState("step_color", func(c *router.Context) error {
		color := c.Update.Text

		// Retrieve previously saved data typed using Generics
		name, _ := router.GetFSMData[string](c, "name")
		age, _ := router.GetFSMData[string](c, "age")

		// Send the final summary
		err := c.Replyf("📝 **Your Profile**\nName: %s\nAge: %s\nColor: %s\n\nProfile saved!", name, age, color)

		// Clear the state AND all the saved data
		c.FSM().Clear()
		return err
	})

	// Fallback for users who are not in any state and type random text
	r.HandleDefault(func(c *router.Context) error {
		if c.Update.Text != "" && !c.Update.IsCommand() {
			return c.Reply("Send /start to begin the questionnaire.")
		}
		return nil
	})

	log.Println("Questionnaire Bot is running. Send /start")
	r.Start(ctx, updatesChannel)
}
