package users

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestService_GetSelf(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/self/get/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.BotInfo{
			Ok:          true,
			ID:          "bot_123",
			DisplayName: "My Bot",
			Login:       "mybot",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	botInfo, err := svc.GetSelf(context.Background())
	if err != nil || botInfo.ID != "bot_123" {
		t.Fatalf("unexpected GetSelf result: %+v, err: %v", botInfo, err)
	}
}

func TestService_GetUserLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/users/getUserLink/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("login") != "user_test" {
			t.Errorf("expected login=user_test, got %s", r.URL.Query().Get("login"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserLinkResponse{
			Ok:       true,
			ChatLink: "https://yandex.ru/chat/user_test",
			CallLink: "https://yandex.ru/call/user_test",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.GetUserLink(context.Background(), GetUserLinkRequest{Login: "user_test"})
	if err != nil || resp.ChatLink == "" {
		t.Fatalf("unexpected GetUserLink result: %+v, err: %v", resp, err)
	}
}
