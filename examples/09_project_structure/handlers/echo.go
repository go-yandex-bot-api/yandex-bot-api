// Package handlers provides example handlers.
package handlers

import (
	"fmt"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

// EchoController demonstrates how to inject dependencies (like a DB) into your handlers.
type EchoController struct {
	Database string // Mock database connection
}

// NewEchoController creates a new echo controller.
func NewEchoController(db string) *EchoController {
	return &EchoController{Database: db}
}

// Register attaches the controller's methods to the router.
func (ctrl *EchoController) Register(r *router.Router) {
	r.HandleCommand("echo", ctrl.handleEcho)
}

func (ctrl *EchoController) handleEcho(c *router.Context) error {
	args := c.Update.CommandArguments()
	if args == "" {
		return c.Reply("Please provide text. Usage: /echo hello")
	}

	// Example of using the injected dependency (ctrl.Database)
	response := fmt.Sprintf("Echoing: %s\n(Logged to: %s)", args, ctrl.Database)
	return c.Reply(response)
}
