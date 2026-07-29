package updates

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig(100)
	if cfg.Offset != 100 || cfg.Limit != 100 || cfg.MinPollInterval != 500*time.Millisecond {
		t.Errorf("unexpected config defaults: %+v", cfg)
	}
}

func TestService_GetUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/getUpdates/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetUpdatesResponse{
			Ok: true,
			Updates: []types.Update{
				{UpdateID: 501, Text: "Update 1"},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	offset := types.UpdateID(500)
	updates, err := svc.GetUpdates(context.Background(), GetUpdatesRequest{}.WithOffset(offset))
	if err != nil || len(updates) != 1 || updates[0].UpdateID != 501 {
		t.Fatalf("unexpected GetUpdates result: %+v, err: %v", updates, err)
	}
}

func TestService_GetUpdatesChannel_WebhookActiveError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.BotInfo{
			Ok:         true,
			WebhookURL: "https://example.com/webhook",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	_, err := svc.GetUpdatesChannel(context.Background(), NewConfig(0))
	if err == nil {
		t.Error("expected error when webhook is active, got nil")
	}
}
