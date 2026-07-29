// Package handlers provides example handlers.
package handlers

import (
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

// RegisterStartHandlers attaches all commands related to onboarding and greeting.
func RegisterStartHandlers(r *router.Router) {
	r.HandleCommand("start", handleStart)
	r.HandleCommand("help", handleHelp)
}

func handleStart(c *router.Context) error {
	// Because c.Bot is available in the context, we can just call c.Reply!
	return c.Reply("Hello! I am a modular bot. Send /echo to test the controller.")
}

func handleHelp(c *router.Context) error {
	return c.Reply("Help section:\n/start - Greeting\n/echo [text] - Echo message")
}
