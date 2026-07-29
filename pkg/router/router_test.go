package router

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/fsm"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestRouter_CommandHandling(t *testing.T) {
	router := NewRouter(nil)

	handled := false
	router.HandleCommand("start", func(_ *Context) error {
		handled = true
		return nil
	})

	update := types.Update{Text: "/start param"}
	if err := router.Process(context.Background(), update); err != nil {
		t.Fatal(err)
	}

	if !handled {
		t.Error("Expected /start command to be handled")
	}
}

func TestRouter_ButtonHandling(t *testing.T) {
	router := NewRouter(nil)

	handled := false
	router.HandleButton("btn_action", func(_ *Context) error {
		handled = true
		return nil
	})

	update := types.Update{
		BotRequest: &types.BotRequest{
			ServerAction: &types.ServerAction{Name: "btn_action"},
		},
	}
	if err := router.Process(context.Background(), update); err != nil {
		t.Fatal(err)
	}

	if !handled {
		t.Error("Expected button action to be handled")
	}
}

func TestRouter_StateHandling(t *testing.T) {
	storage := fsm.NewMemoryStorage(time.Hour)
	router := NewRouter(nil).WithStorage(storage)

	// Create a fake update from a user
	update := types.Update{
		Text: "25",
		From: &types.Sender{Login: "user_login_1"},
	}

	// Set state manually using router helper
	router.SetState(update, "WAITING_AGE")

	handled := false
	router.HandleState("WAITING_AGE", func(c *Context) error {
		handled = true
		// Clear state after handling
		router.ClearState(c.Update)
		return nil
	})

	if err := router.Process(context.Background(), update); err != nil {
		t.Fatal(err)
	}

	if !handled {
		t.Error("Expected FSM state to be handled")
	}

	// Verify state was cleared
	if state := storage.Get("user_login_1"); state != "" {
		t.Errorf("Expected FSM state to be empty, got %s", state)
	}
}

func TestRouter_TextFallback(t *testing.T) {
	router := NewRouter(nil)

	handled := false
	router.HandleText(func(_ *Context) error {
		handled = true
		return nil
	})

	update := types.Update{Text: "Just a regular text without slash"}
	if err := router.Process(context.Background(), update); err != nil {
		t.Fatal(err)
	}

	if !handled {
		t.Error("Expected fallback text handler to be executed")
	}
}

func TestRouter_Middleware(t *testing.T) {
	router := NewRouter(nil)

	middleware1Executed := false
	middleware2Executed := false
	handlerExecuted := false

	router.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			middleware1Executed = true
			return next(c)
		}
	})

	router.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			middleware2Executed = true
			return next(c)
		}
	})

	router.HandleText(func(_ *Context) error {
		handlerExecuted = true
		return nil
	})

	if err := router.Process(context.Background(), types.Update{Text: "Test middleware"}); err != nil {
		t.Fatal(err)
	}

	if !middleware1Executed {
		t.Error("Expected middleware 1 to execute")
	}
	if !middleware2Executed {
		t.Error("Expected middleware 2 to execute")
	}
	if !handlerExecuted {
		t.Error("Expected handler to execute")
	}
}

func TestRouter_RecoverMiddleware(t *testing.T) {
	router := NewRouter(nil)
	router.Use(RecoverMiddleware())

	router.HandleText(func(_ *Context) error {
		panic("something went terribly wrong")
	})

	err := router.Process(context.Background(), types.Update{Text: "trigger panic"})
	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
}

func TestRouter_OnError(t *testing.T) {
	router := NewRouter(nil)

	customErr := errors.New("custom error")
	errCh := make(chan error, 1)

	router.OnError(func(c *Context, err error) {
		errCh <- err
	})

	router.HandleText(func(_ *Context) error {
		return customErr
	})

	ch := make(chan types.Update, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch <- types.Update{Text: "error test"}

	go router.Start(ctx, ch)

	select {
	case captured := <-errCh:
		if !errors.Is(captured, customErr) {
			t.Errorf("expected OnError to capture customErr, got %v", captured)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for OnError to be called")
	}
}

func TestContext_Helpers(t *testing.T) {
	u := types.Update{
		From: &types.Sender{Login: "admin_user"},
		Chat: &types.Chat{ID: "chat_999"},
	}

	c := &Context{Update: u}

	if c.SenderLogin() != "admin_user" {
		t.Errorf("expected SenderLogin admin_user, got %s", c.SenderLogin())
	}
	if !c.SenderLoginIs("admin_user") || c.SenderLoginIs("other") {
		t.Error("SenderLoginIs validation failed")
	}
	if c.ChatID() != "chat_999" {
		t.Errorf("expected ChatID chat_999, got %s", c.ChatID())
	}
}
