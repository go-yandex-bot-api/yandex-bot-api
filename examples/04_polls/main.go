// Package main provides an example.
package main

import (
	"context"
	"fmt"
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

	// Command to create a poll
	r.HandleCommand("poll", func(c *router.Context) error {
		u := c.Update
		req := yabotapi.CreatePollRequest{
			ChatID:  u.GetChatID(),
			Title:   "What is your favorite programming language?",
			Answers: []string{"Go", "Python", "Rust", "Java"},
		}
		if req.ChatID == "" {
			req.Login = u.GetFromLogin()
		}

		resp, err := bot.Polls.Create(c.Ctx, req)
		if err != nil {
			log.Println("Error sending poll:", err)
			return err
		}

		msg := yabotapi.NewReply(&u, fmt.Sprintf("Poll created! Message ID: %v", resp.MessageID))
		// You can save resp.MessageID to fetch results later (bot.Polls.GetResults)
		log.Printf("Poll created, ID: %v", resp.MessageID)
		if _, err := bot.Messages.SendText(c.Ctx, msg); err != nil {
			return err
		}
		return nil
	})

	// Handle vote events in real-time
	r.HandleVote(func(c *router.Context) error {
		vote := c.Update.Vote
		log.Printf("User %s voted in poll %s for options: %v", c.SenderLogin(), vote.PollID, vote.OptionIDs)
		return c.Replyf("Thank you for voting in poll %s!", vote.PollID)
	})

	log.Println("Bot is running. Send '/poll' to create a poll.")
	r.Start(ctx, updatesChannel)
}
