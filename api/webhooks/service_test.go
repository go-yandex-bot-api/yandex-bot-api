package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestService_SetWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/self/update/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client, false)

	if err := svc.SetWebhook(context.Background(), "https://example.com/webhook"); err != nil {
		t.Fatalf("SetWebhook error: %v", err)
	}

	if err := svc.SetWebhook(context.Background(), ""); err != nil {
		t.Fatalf("SetWebhook clear error: %v", err)
	}
}

func TestService_ListenForWebhook(t *testing.T) {
	svc := NewService(nil, true)
	ctx := context.Background()
	ch, handler := svc.ListenForWebhook(ctx)

	payload := map[string]interface{}{
		"updates": []types.Update{
			{UpdateID: 1001, Text: "hello webhook"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	select {
	case u := <-ch:
		if u.UpdateID != 1001 || u.Text != "hello webhook" {
			t.Errorf("unexpected update received: %+v", u)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for update from channel")
	}
}

func TestService_ListenForWebhook_MethodNotAllowed(t *testing.T) {
	svc := NewService(nil, false)
	_, handler := svc.ListenForWebhook(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/webhook", http.NoBody)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}
