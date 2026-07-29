package yabotapi

import (
	"errors"
	"strings"

	"github.com/go-yandex-bot-api/yandex-bot-api/api/chats"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/files"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/messages"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/polls"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/updates"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/users"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/webhooks"
	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Bot is the main application client that aggregates the core HTTP client and all domain services.
type Bot struct {
	Client *core.Client

	// Services
	Chats    *chats.Service
	Messages *messages.Service
	Polls    *polls.Service
	Updates  *updates.Service
	Files    *files.Service
	Webhooks *webhooks.Service
	Users    *users.Service
}

// APIError represents an error returned by the Yandex API.
type APIError = core.APIError

// Update represents a single incoming event from the server.
type Update = types.Update

// Chat represents a chat or channel in Yandex Messenger.
type Chat = types.Chat

// UserLogin represents a user's login identifier.
type UserLogin = types.UserLogin

// ChatID represents a chat's unique identifier.
type ChatID = types.ChatID

// UpdateID represents an update's unique identifier.
type UpdateID = types.UpdateID

// ThreadID represents a thread's unique identifier.
type ThreadID = types.ThreadID

// MessageID represents a message's unique identifier.
type MessageID = types.MessageID

// Option configures the Bot's underlying HTTP client at construction time.
type Option = core.Option

// ErrorHandlingConfig defines rules for retrying failed requests.
type ErrorHandlingConfig = core.ErrorHandlingConfig

// HTTPClient is an interface representing an HTTP client.
type HTTPClient = core.HTTPClient

// WithClient sets the underlying HTTP client.
func WithClient(c HTTPClient) Option { return core.WithClient(c) }

// WithAPIURL sets the base URL for the Yandex API.
func WithAPIURL(u string) Option { return core.WithAPIURL(u) }

// WithMaxRetries sets the maximum number of retries for failed requests.
func WithMaxRetries(r int) Option { return core.WithMaxRetries(r) }

// WithDebug enables or disables debug logging.
func WithDebug(d bool) Option { return core.WithDebug(d) }

// WithErrorHandlingConfig sets the error handling configuration.
func WithErrorHandlingConfig(cfg *ErrorHandlingConfig) Option {
	return core.WithErrorHandlingConfig(cfg)
}

// Common API Sentinel errors for use with errors.Is.
var (
	ErrBadRequest   = core.ErrBadRequest
	ErrUnauthorized = core.ErrUnauthorized
	ErrForbidden    = core.ErrForbidden
	ErrNotFound     = core.ErrNotFound
	ErrConflict     = core.ErrConflict
	ErrRateLimited  = core.ErrRateLimited
	ErrServerError  = core.ErrServerError
)

// NewBot creates a new Bot instance with default settings, applies options, and initializes its services.
func NewBot(token string, opts ...Option) (*Bot, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("bot token cannot be empty")
	}
	client := core.NewClient(token, opts...)
	b := &Bot{
		Client: client,
	}

	b.initServices()
	return b, nil
}

func (b *Bot) initServices() {
	b.Chats = chats.NewService(b.Client)
	b.Messages = messages.NewService(b.Client)
	b.Polls = polls.NewService(b.Client)
	b.Updates = updates.NewService(b.Client)
	b.Files = files.NewService(b.Client)
	b.Webhooks = webhooks.NewService(b.Client, b.Client.Debug())
	b.Users = users.NewService(b.Client)
}
