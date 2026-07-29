package messages

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestService_SendText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendText/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req SendTextRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Text != "Hello World" || req.ChatID != "chat_123" {
			t.Errorf("unexpected request payload: %+v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 101,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.SendText(context.Background(), SendTextRequest{
		ChatID: "chat_123",
		Text:   "Hello World",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MessageID != 101 {
		t.Errorf("expected MessageID 101, got %d", resp.MessageID)
	}
}

func TestService_SendText_ValidationError(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.SendText(context.Background(), SendTextRequest{})
	if err == nil {
		t.Error("expected error for empty ChatID and Login, got nil")
	}

	_, err = svc.SendText(context.Background(), SendTextRequest{ChatID: "c1"})
	if err == nil {
		t.Error("expected error for empty Text, got nil")
	}
}

func TestService_SendSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendSystemMessage/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 102,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.SendSystemMessage(context.Background(), SendSystemMessageRequest{
		ChatID: "chat_123",
		Text:   "System notice",
	})
	if err != nil || resp.MessageID != 102 {
		t.Fatalf("unexpected response or error: %v", err)
	}
}

func TestService_SendSticker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendSticker/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 103,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.SendSticker(context.Background(), SendStickerRequest{
		ChatID:       "chat_123",
		StickerSetID: "set1",
		StickerID:    "sticker1",
	})
	if err != nil || resp.MessageID != 103 {
		t.Fatalf("unexpected response or error: %v", err)
	}
}

func TestService_SendTyping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendTyping/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	err := svc.SendTyping(context.Background(), SendTypingRequest{ChatID: "chat_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Delete_Pin_Unpin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	ctx := context.Background()
	if err := svc.Delete(ctx, DeleteMessageRequest{ChatID: "c1", MessageID: 1}); err != nil {
		t.Errorf("Delete error: %v", err)
	}
	if err := svc.Pin(ctx, PinMessageRequest{ChatID: "c1", MessageID: 1}); err != nil {
		t.Errorf("Pin error: %v", err)
	}
	if err := svc.Unpin(ctx, UnpinMessageRequest{ChatID: "c1", MessageID: 1}); err != nil {
		t.Errorf("Unpin error: %v", err)
	}
}
