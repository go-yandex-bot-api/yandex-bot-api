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

func TestContext_EditHelpers(t *testing.T) {
	u := types.Update{
		MessageID: 101,
		From:      &types.Sender{Login: "admin_user"},
		Chat:      &types.Chat{ID: "chat_999"},
	}

	c := &Context{Update: u}

	if c.MessageID() != 101 {
		t.Errorf("expected MessageID 101, got %d", c.MessageID())
	}

	// Without Bot initialized, methods should fail gracefully
	if err := c.EditMessage(101, "new text"); err == nil {
		t.Error("expected error when Bot is nil in EditMessage")
	}
	if err := c.EditCurrentMessage("new text"); err == nil {
		t.Error("expected error when Bot is nil in EditCurrentMessage")
	}
	if err := c.EditCurrentMessageWithKeyboard("new text", nil); err == nil {
		t.Error("expected error when Bot is nil in EditCurrentMessageWithKeyboard")
	}

	// Test zero message_id error
	emptyContext := &Context{Update: types.Update{}}
	if err := emptyContext.EditCurrentMessage("new text"); err == nil {
		t.Error("expected error for empty message_id in EditCurrentMessage")
	}
}

func TestContext_BindPayload(t *testing.T) {
	type SamplePayload struct {
		ItemID int    `json:"item_id"`
		Page   int    `json:"page"`
		Action string `json:"action,omitempty"`
	}

	makeCtx := func(raw []byte) *Context {
		if raw == nil {
			return &Context{Update: types.Update{}}
		}
		return &Context{
			Update: types.Update{
				BotRequest: &types.BotRequest{
					ServerAction: &types.ServerAction{
						Name:    "test_action",
						Payload: raw,
					},
				},
			},
		}
	}

	// 1. Desktop Struct: native JSON object
	t.Run("Desktop Struct", func(t *testing.T) {
		ctx := makeCtx([]byte(`{"item_id": 123, "page": 2}`))
		var p SamplePayload
		if err := ctx.BindPayload(&p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ItemID != 123 || p.Page != 2 {
			t.Errorf("expected {123, 2}, got %+v", p)
		}
	})

	// 2. Mobile Struct: string-escaped JSON object
	t.Run("Mobile Struct String-Escaped", func(t *testing.T) {
		ctx := makeCtx([]byte(`"{\"item_id\": 123, \"page\": 2, \"action\": \"buy\"}"`))
		var p SamplePayload
		if err := ctx.BindPayload(&p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ItemID != 123 || p.Page != 2 || p.Action != "buy" {
			t.Errorf("expected {123, 2, buy}, got %+v", p)
		}
	})

	// 3. Desktop Primitive: {"value": 42} -> int
	t.Run("Desktop Primitive Int", func(t *testing.T) {
		ctx := makeCtx([]byte(`{"value": 42}`))
		var val int
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	// 4. Mobile Primitive: "{\"value\": 42}" -> int
	t.Run("Mobile Primitive Int String-Escaped", func(t *testing.T) {
		ctx := makeCtx([]byte(`"{\"value\": 42}"`))
		var val int
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	// 5. Desktop Primitive String: {"value": "admin"} -> string
	t.Run("Desktop Primitive String", func(t *testing.T) {
		ctx := makeCtx([]byte(`{"value": "admin"}`))
		var val string
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "admin" {
			t.Errorf("expected admin, got %s", val)
		}
	})

	// 6. Mobile Primitive String: "{\"value\": \"admin\"}" -> string
	t.Run("Mobile Primitive String String-Escaped", func(t *testing.T) {
		ctx := makeCtx([]byte(`"{\"value\": \"admin\"}"`))
		var val string
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "admin" {
			t.Errorf("expected admin, got %s", val)
		}
	})

	// 7. Raw String Direct: "custom_str" -> string
	t.Run("Raw String Direct", func(t *testing.T) {
		ctx := makeCtx([]byte(`"custom_str"`))
		var val string
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "custom_str" {
			t.Errorf("expected custom_str, got %s", val)
		}
	})

	// 8. Mobile Array / Slice: "[10, 20, 30]" -> []int
	t.Run("Mobile Array String-Escaped", func(t *testing.T) {
		ctx := makeCtx([]byte(`"[10, 20, 30]"`))
		var val []int
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(val) != 3 || val[0] != 10 || val[2] != 30 {
			t.Errorf("expected [10 20 30], got %v", val)
		}
	})

	// 9. Empty / Nil / Null Payload errors
	t.Run("Empty and Null Payload", func(t *testing.T) {
		nilCtx := makeCtx(nil)
		var p SamplePayload
		if err := nilCtx.BindPayload(&p); err == nil {
			t.Error("expected error for nil payload")
		}

		emptyCtx := makeCtx([]byte(""))
		if err := emptyCtx.BindPayload(&p); err == nil {
			t.Error("expected error for empty payload")
		}

		nullCtx := makeCtx([]byte("null"))
		if err := nullCtx.BindPayload(&p); err == nil {
			t.Error("expected error for 'null' payload")
		}
	})

	// 10. Yandex Float Integer in Struct: 101.0 -> int 101
	t.Run("Yandex Float Integer in Struct", func(t *testing.T) {
		type Product struct {
			ItemID int     `json:"item_id"`
			Price  float64 `json:"price"`
			Name   string  `json:"name"`
		}
		raw := []byte(`{"price": 99.5, "item_id": 101.0, "name": "version 1.0"}`)
		ctx := makeCtx(raw)
		var p Product
		if err := ctx.BindPayload(&p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ItemID != 101 || p.Price != 99.5 || p.Name != "version 1.0" {
			t.Errorf("expected {101, 99.5, 'version 1.0'}, got %+v", p)
		}
	})

	// 11. Yandex Float Integer in Primitive: {"value": 42.0} -> int 42
	t.Run("Yandex Float Integer in Primitive", func(t *testing.T) {
		ctx := makeCtx([]byte(`{"value": 42.0}`))
		var val int
		if err := ctx.BindPayload(&val); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	// 12. Preserve Fractional Floats: 5.6, 5.06, 0.007
	t.Run("Preserve Fractional Floats", func(t *testing.T) {
		type Metrics struct {
			Rate1 float64 `json:"rate1"`
			Rate2 float64 `json:"rate2"`
			Rate3 float64 `json:"rate3"`
			Count int     `json:"count"`
		}
		raw := []byte(`{"rate1": 5.6, "rate2": 5.06, "rate3": 0.007, "count": 10.0}`)
		ctx := makeCtx(raw)
		var m Metrics
		if err := ctx.BindPayload(&m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.Rate1 != 5.6 || m.Rate2 != 5.06 || m.Rate3 != 0.007 || m.Count != 10 {
			t.Errorf("expected {5.6, 5.06, 0.007, 10}, got %+v", m)
		}

		// Also check float primitive {"value": 5.06}
		ctxFloat := makeCtx([]byte(`{"value": 5.06}`))
		var fVal float64
		if err := ctxFloat.BindPayload(&fVal); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fVal != 5.06 {
			t.Errorf("expected 5.06, got %f", fVal)
		}
	})
}
