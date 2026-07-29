package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/messages"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/updates"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/fsm"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Context wraps an incoming Update and provides helper methods.
type Context struct {
	Ctx    context.Context //nolint:containedctx // Needed for request scoped data
	Update types.Update
	Bot    *yabotapi.Bot
	router *Router
}

// ContextFSM provides a scoped FSM interface for the current user.
type ContextFSM struct {
	storage  fsm.Storage
	senderID string
}

// SetState updates the FSM state for the current user.
func (f *ContextFSM) SetState(state string) {
	if f.storage != nil && f.senderID != "" {
		f.storage.Set(f.senderID, state)
	}
}

// GetState retrieves the FSM state for the current user.
func (f *ContextFSM) GetState() string {
	if f.storage != nil && f.senderID != "" {
		return f.storage.Get(f.senderID)
	}
	return ""
}

// Clear removes the FSM state and payload data for the current user.
func (f *ContextFSM) Clear() {
	if f.storage != nil && f.senderID != "" {
		f.storage.Delete(f.senderID)
	}
}

// SetData saves an arbitrary key-value pair for the current user.
func (f *ContextFSM) SetData(key string, value any) {
	if f.storage != nil && f.senderID != "" {
		f.storage.SetData(f.senderID, key, value)
	}
}

// GetData retrieves an arbitrary key-value pair for the current user.
func (f *ContextFSM) GetData(key string) (any, bool) {
	if f.storage != nil && f.senderID != "" {
		return f.storage.GetData(f.senderID, key)
	}
	return nil, false
}

// GetFSMData retrieves FSM payload typed directly using Generics, avoiding manual type assertion.
func GetFSMData[T any](c *Context, key string) (T, bool) {
	var zero T
	val, ok := c.FSM().GetData(key)
	if !ok || val == nil {
		return zero, false
	}
	typedVal, ok := val.(T)
	if !ok {
		return zero, false
	}
	return typedVal, true
}

// FSM returns a scoped state machine manager for the current context.
func (c *Context) FSM() *ContextFSM {
	c.router.mu.RLock()
	defer c.router.mu.RUnlock()
	return &ContextFSM{
		storage:  c.router.storage,
		senderID: c.router.GetSenderID(c.Update),
	}
}

// Reply sends a text message back to the sender of the update.
func (c *Context) Reply(text string) error {
	if c.Bot == nil {
		return errors.New("bot is not set in context")
	}

	var req messages.SendTextRequest
	req.Text = text
	req.ThreadID = c.Update.ThreadID
	req.ReplyMessageID = c.Update.MessageID

	if c.Update.Chat != nil && c.Update.Chat.ID != "" {
		req.ChatID = c.Update.Chat.ID
	} else if login := c.Update.GetFromLogin(); login != "" {
		req.Login = login
	} else {
		return errors.New("cannot determine recipient for reply")
	}

	_, err := c.Bot.Messages.SendText(c.Ctx, req)
	return err
}

// Replyf formats a string according to a format specifier and sends it back to the sender.
func (c *Context) Replyf(format string, args ...any) error {
	return c.Reply(fmt.Sprintf(format, args...))
}

// Send sends a new text message to the current chat/user without replying/quoting the original message.
func (c *Context) Send(text string) error {
	if c.Bot == nil {
		return errors.New("bot is not set in context")
	}

	var req messages.SendTextRequest
	req.Text = text

	if c.Update.Chat != nil && c.Update.Chat.ID != "" {
		req.ChatID = c.Update.Chat.ID
	} else if login := c.Update.GetFromLogin(); login != "" {
		req.Login = login
	} else {
		return errors.New("cannot determine recipient for send")
	}

	_, err := c.Bot.Messages.SendText(c.Ctx, req)
	return err
}

// Sendf formats a string and sends a new text message to the current chat/user without quoting.
func (c *Context) Sendf(format string, args ...any) error {
	return c.Send(fmt.Sprintf(format, args...))
}

// SendWithKeyboard sends a new text message with a keyboard without quoting the original message.
func (c *Context) SendWithKeyboard(text string, keyboard *types.SuggestButtons) error {
	if c.Bot == nil {
		return errors.New("bot is not set in context")
	}

	var req messages.SendTextRequest
	req.Text = text
	req.SuggestButtons = keyboard

	if c.Update.Chat != nil && c.Update.Chat.ID != "" {
		req.ChatID = c.Update.Chat.ID
	} else if login := c.Update.GetFromLogin(); login != "" {
		req.Login = login
	} else {
		return errors.New("cannot determine recipient for send")
	}

	_, err := c.Bot.Messages.SendText(c.Ctx, req)
	return err
}

// SendWithKeyboardf formats a string and sends a new message with a keyboard without quoting.
func (c *Context) SendWithKeyboardf(keyboard *types.SuggestButtons, format string, args ...any) error {
	return c.SendWithKeyboard(fmt.Sprintf(format, args...), keyboard)
}

// SendTo sends a new text message to a specific ChatID or UserLogin string.
func (c *Context) SendTo(recipient, text string) error {
	if c.Bot == nil {
		return errors.New("bot is not set in context")
	}
	if recipient == "" {
		return errors.New("recipient cannot be empty")
	}

	var req messages.SendTextRequest
	req.Text = text

	// Check if recipient is a chat ID UUID or a user login
	if len(recipient) > 30 && (recipient[8] == '-' || (len(recipient) > 36 && recipient[36] == '_')) {
		req.ChatID = types.ChatID(recipient)
	} else {
		req.Login = types.UserLogin(recipient)
	}

	_, err := c.Bot.Messages.SendText(c.Ctx, req)
	return err
}

// SenderLogin returns the login of the update sender safely.
func (c *Context) SenderLogin() types.UserLogin {
	return c.Update.GetFromLogin()
}

// ChatID returns the ChatID of the update safely.
func (c *Context) ChatID() types.ChatID {
	return c.Update.GetChatID()
}

// SenderLoginIs checks if the sender's login matches targetLogin without explicit string conversion.
func (c *Context) SenderLoginIs(targetLogin string) bool {
	return string(c.Update.GetFromLogin()) == targetLogin
}

// IsFile returns true if the update contains a document file attachment.
func (c *Context) IsFile() bool {
	return c.Update.File != nil
}

// IsImage returns true if the update contains image attachments.
func (c *Context) IsImage() bool {
	return len(c.Update.Images) > 0
}

// IsSticker returns true if the update contains a sticker.
func (c *Context) IsSticker() bool {
	return c.Update.Sticker != nil
}

// IsVote returns true if the update contains a poll vote event.
func (c *Context) IsVote() bool {
	return c.Update.Vote != nil
}

// IsPoll returns true if the update contains a poll object.
func (c *Context) IsPoll() bool {
	return c.Update.Poll != nil
}

// FirstImage returns the first image attachment in highest resolution (original) if available.
func (c *Context) FirstImage() (*types.Image, bool) {
	if len(c.Update.Images) > 0 && len(c.Update.Images[0]) > 0 {
		lastIdx := len(c.Update.Images[0]) - 1
		return &c.Update.Images[0][lastIdx], true
	}
	return nil, false
}

// FirstFile returns the file attachment if available.
func (c *Context) FirstFile() (*types.File, bool) {
	if c.Update.File != nil {
		return c.Update.File, true
	}
	return nil, false
}

// BindPayload unmarshals the payload from an ActionButton click into dest.
// It supports both structs/maps and primitive target pointers (e.g. &int, &string).
func (c *Context) BindPayload(dest interface{}) error {
	if c.Update.BotRequest == nil || c.Update.BotRequest.ServerAction == nil ||
		c.Update.BotRequest.ServerAction.Payload == nil {
		return errors.New("no payload to bind")
	}

	payloadRaw := c.Update.BotRequest.ServerAction.Payload

	// 1. Try unmarshaling directly into dest
	if err := json.Unmarshal(payloadRaw, dest); err == nil {
		return nil
	}

	// 2. If payload was normalized into {"value": ...}, try extracting "value"
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(payloadRaw, &wrapper); err == nil {
		if val, exists := wrapper["value"]; exists {
			return json.Unmarshal(val, dest)
		}
	}

	return json.Unmarshal(payloadRaw, dest)
}

// HandlerFunc is the signature for functions that process an Update.
type HandlerFunc func(c *Context) error

// MiddlewareFunc is a function that wraps a HandlerFunc to execute logic before/after it.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type stateRegexpHandler struct {
	pattern *regexp.Regexp
	handler HandlerFunc
}

// Router is a simple event dispatcher for incoming updates.
type Router struct {
	mu                  sync.RWMutex
	bot                 *yabotapi.Bot
	commandHandlers     map[string]HandlerFunc
	buttonHandlers      map[string]HandlerFunc
	stateHandlers       map[string]HandlerFunc
	stateRegexpHandlers []stateRegexpHandler
	textHandler         HandlerFunc
	fileHandler         HandlerFunc
	imageHandler        HandlerFunc
	stickerHandler      HandlerFunc
	voteHandler         HandlerFunc
	pollHandler         HandlerFunc
	defaultHandler      HandlerFunc
	errorHandler        func(c *Context, err error)
	middlewares         []MiddlewareFunc
	storage             fsm.Storage
	chain               HandlerFunc
}

// OnError sets a global error handler for any uncaught errors returned by handlers.
func (r *Router) OnError(h func(c *Context, err error)) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorHandler = h
	return r
}

// NewRouter creates a new instance of Router without FSM storage.
func NewRouter(bot *yabotapi.Bot) *Router {
	return &Router{
		bot:             bot,
		commandHandlers: make(map[string]HandlerFunc),
		buttonHandlers:  make(map[string]HandlerFunc),
		stateHandlers:   make(map[string]HandlerFunc),
	}
}

// WithStorage attaches an FSM Storage (like MemoryStorage) to the Router.
func (r *Router) WithStorage(storage fsm.Storage) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storage = storage
	return r
}

// Use adds one or more middlewares to the router.
func (r *Router) Use(mw ...MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, mw...)
	r.chain = nil
}

// HandleCommand registers a handler for a specific command (e.g., "start" for "/start").
func (r *Router) HandleCommand(command string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commandHandlers[command] = handler
	r.chain = nil
}

// HandleButton registers a handler for a specific ServerAction Name (when a user clicks an ActionButton).
func (r *Router) HandleButton(actionName string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buttonHandlers[actionName] = handler
	r.chain = nil
}

// HandleText registers a fallback handler for any text message that is not a command.
func (r *Router) HandleText(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.textHandler = handler
	r.chain = nil
}

// HandleState registers a handler for updates coming from users in a specific FSM state.
func (r *Router) HandleState(state string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stateHandlers[state] = handler
	r.chain = nil
}

// HandleStateRegexp registers a handler for updates coming from users in an FSM state that matches the pattern.
// Returns an error if the pattern is invalid.
func (r *Router) HandleStateRegexp(pattern string, handler HandlerFunc) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid state regexp pattern %q: %w", pattern, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.stateRegexpHandlers = append(r.stateRegexpHandlers, stateRegexpHandler{
		pattern: re,
		handler: handler,
	})
	r.chain = nil
	return nil
}

// HandleFile registers a handler for updates that contain a file attachment.
func (r *Router) HandleFile(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fileHandler = handler
	r.chain = nil
}

// HandleImage registers a handler for updates that contain images.
func (r *Router) HandleImage(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.imageHandler = handler
	r.chain = nil
}

// HandleSticker registers a handler for updates that contain stickers.
func (r *Router) HandleSticker(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stickerHandler = handler
	r.chain = nil
}

// HandleVote registers a handler for poll vote events.
func (r *Router) HandleVote(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.voteHandler = handler
	r.chain = nil
}

// HandlePoll registers a handler for updates that contain a poll creation/object.
func (r *Router) HandlePoll(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pollHandler = handler
	r.chain = nil
}

// HandleDefault registers a fallback handler for any update that wasn't caught by other handlers.
func (r *Router) HandleDefault(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = handler
	r.chain = nil
}

// GetSenderID determines a unique ID for the sender of the update (used for FSM).
func (r *Router) GetSenderID(u types.Update) string {
	chatID := string(u.GetChatID())
	login := string(u.GetFromLogin())
	if chatID != "" && login != "" {
		return chatID + ":" + login
	}
	if login != "" {
		return login
	}
	return chatID
}

// SetState updates the FSM state for the sender of this update.
func (r *Router) SetState(u types.Update, state string) {
	r.mu.RLock()
	st := r.storage
	r.mu.RUnlock()
	if st == nil {
		return // Ignore if no storage configured
	}
	senderID := r.GetSenderID(u)
	if senderID != "" {
		st.Set(senderID, state)
	}
}

// Router returns the attached Router instance.
func (c *Context) Router() *Router { return c.router }

// ReplyWithKeyboard sends a text message with a keyboard back to the sender.
func (c *Context) ReplyWithKeyboard(text string, keyboard *types.SuggestButtons) error {
	if c.Bot == nil {
		return errors.New("bot is not set in context")
	}

	var req messages.SendTextRequest
	req.Text = text
	req.SuggestButtons = keyboard
	req.ThreadID = c.Update.ThreadID
	req.ReplyMessageID = c.Update.MessageID

	if c.Update.Chat != nil && c.Update.Chat.ID != "" {
		req.ChatID = c.Update.Chat.ID
	} else if login := c.Update.GetFromLogin(); login != "" {
		req.Login = login
	} else {
		return errors.New("cannot determine recipient for reply")
	}

	_, err := c.Bot.Messages.SendText(c.Ctx, req)
	return err
}

// ReplyWithKeyboardf formats a string according to a format specifier and sends it with a keyboard.
func (c *Context) ReplyWithKeyboardf(keyboard *types.SuggestButtons, format string, args ...any) error {
	return c.ReplyWithKeyboard(fmt.Sprintf(format, args...), keyboard)
}

// ClearState clears the FSM state for the sender of this update.
func (r *Router) ClearState(u types.Update) {
	r.mu.RLock()
	st := r.storage
	r.mu.RUnlock()
	if st == nil {
		return
	}
	senderID := r.GetSenderID(u)
	if senderID != "" {
		st.Delete(senderID)
	}
}

// buildChain precompiles the middleware chain.
func (r *Router) buildChain() HandlerFunc {
	chain := func(c *Context) error {
		if handler := r.resolveHandler(c.Update); handler != nil {
			return handler(c)
		}
		return nil
	}

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		chain = r.middlewares[i](chain)
	}

	return chain
}

// Process routes a single update to the appropriate handler.
func (r *Router) Process(ctx context.Context, u types.Update) error {
	r.mu.RLock()
	chain := r.chain
	r.mu.RUnlock()

	if chain == nil {
		r.mu.Lock()
		if r.chain == nil {
			r.chain = r.buildChain()
		}
		chain = r.chain
		r.mu.Unlock()
	}

	c := &Context{
		Ctx:    ctx,
		Update: u,
		Bot:    r.bot,
		router: r,
	}
	return chain(c)
}

// resolveHandler finds the target handler for an update without executing middlewares.
//nolint:gocyclo // complex routing logic intentionally handles many update types
func (r *Router) resolveHandler(u types.Update) HandlerFunc {
	r.mu.RLock()

	if u.BotRequest != nil && u.BotRequest.ServerAction != nil {
		if handler, exists := r.buttonHandlers[u.BotRequest.ServerAction.Name]; exists {
			r.mu.RUnlock()
			return handler
		}
	}

	if u.IsCommand() { // command routing logic
		if handler, exists := r.commandHandlers[u.Command()]; exists {
			r.mu.RUnlock()
			return handler
		}
	}

	storage := r.storage
	r.mu.RUnlock()

	if storage != nil { //nolint:nestif // state routing logic
		senderID := r.GetSenderID(u)
		if senderID != "" {
			state := storage.Get(senderID)
			if state != "" {
				r.mu.RLock()
				if handler, exists := r.stateHandlers[state]; exists {
					r.mu.RUnlock()
					return handler
				}
				for _, sh := range r.stateRegexpHandlers {
					if sh.pattern.MatchString(state) {
						r.mu.RUnlock()
						return sh.handler
					}
				}
				r.mu.RUnlock()
			}
		}
	}

	r.mu.RLock()
	if u.Vote != nil && r.voteHandler != nil {
		handler := r.voteHandler
		r.mu.RUnlock()
		return handler
	}

	if u.Poll != nil && r.pollHandler != nil {
		handler := r.pollHandler
		r.mu.RUnlock()
		return handler
	}

	if u.File != nil && r.fileHandler != nil {
		handler := r.fileHandler
		r.mu.RUnlock()
		return handler
	}

	if len(u.Images) > 0 && r.imageHandler != nil {
		handler := r.imageHandler
		r.mu.RUnlock()
		return handler
	}

	if u.Sticker != nil && r.stickerHandler != nil {
		handler := r.stickerHandler
		r.mu.RUnlock()
		return handler
	}

	if u.Text != "" && !u.IsCommand() {
		if r.textHandler != nil {
			r.mu.RUnlock()
			return r.textHandler
		}
	}

	handler := r.defaultHandler
	r.mu.RUnlock()
	return handler
}

// StartPolling acquires updates channel and dispatches events until context cancellation.
func (r *Router) StartPolling(ctx context.Context, offset types.UpdateID) error {
	ch, err := r.bot.Updates.GetUpdatesChannel(ctx, updates.NewConfig(offset))
	if err != nil {
		return err
	}
	r.Start(ctx, ch)
	return nil
}

// Start consumes the updates channel and dispatches each update concurrently in a new goroutine.
// Handler errors are logged if they are not handled by a middleware.
// A semaphore limits concurrent handler goroutines to prevent resource exhaustion.
func (r *Router) Start(ctx context.Context, ch <-chan types.Update) {
	const maxConcurrency = 50
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	dispatch := func(u types.Update) {
		wg.Add(1)
		go func(update types.Update) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("[Router] panic recovered processing update %d: %v", update.UpdateID, rec)
				}
			}()

			// 30 seconds is the context timeout per update processing
			reqCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second) //nolint:mnd // hard cap per update
			defer cancel()

			if err := r.Process(reqCtx, update); err != nil {
				log.Printf("[Router] unhandled error processing update %d: %v", update.UpdateID, err)
				r.mu.RLock()
				errH := r.errorHandler
				r.mu.RUnlock()
				if errH != nil {
					c := &Context{
						Bot:    r.bot,
						Update: update,
						Ctx:    reqCtx,
						router: r,
					}
					errH(c, err)
				}
			}
		}(u)
	}

	for {
		select {
		case <-ctx.Done():
		drainLoop:
			for {
				select {
				case update, ok := <-ch:
					if !ok {
						break drainLoop
					}
					dispatch(update)
				default:
					break drainLoop
				}
			}
			wg.Wait()
			return
		case update, ok := <-ch:
			if !ok {
				wg.Wait()
				return
			}
			dispatch(update)
		}
	}
}
