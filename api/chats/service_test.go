package chats

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestService_CreateChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/chats/create/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req CreateChatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "Test Group" {
			t.Errorf("expected name 'Test Group', got '%s'", req.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CreateChatResponse{
			Ok:         true,
			ChatID:     "chat_created_123",
			InviteLink: "https://yandex.ru/chat/123",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.CreateChat(context.Background(), CreateChatRequest{
		Name: "Test Group",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ChatID != "chat_created_123" {
		t.Errorf("expected ChatID 'chat_created_123', got '%s'", resp.ChatID)
	}
}

func TestService_CreateChat_Validation(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.CreateChat(context.Background(), CreateChatRequest{})
	if err == nil {
		t.Error("expected error for empty Name, got nil")
	}
}

func TestService_GetChats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/chats/get/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("expected limit=5 query param, got %s", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetChatsResponse{
			Ok:    true,
			Chats: []types.Chat{{ID: "c1", Title: "Chat 1"}},
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	chats, err := svc.GetChats(context.Background(), GetChatsRequest{}.WithLimit(5))
	if err != nil {
		t.Fatalf("GetChats error: %v", err)
	}
	if len(chats) != 1 || chats[0].ID != "c1" {
		t.Errorf("unexpected chats result: %+v", chats)
	}
}

func TestService_GetMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/chats/getMembers/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("chat_id") != "chat1" {
			t.Errorf("expected chat_id=chat1, got %s", r.URL.Query().Get("chat_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetMembersResponse{
			Ok:      true,
			Members: []ChatMember{{Login: "user1", Role: "admin"}},
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	members, err := svc.GetMembers(context.Background(), GetMembersRequest{ChatID: "chat1"})
	if err != nil || len(members) != 1 {
		t.Fatalf("unexpected members result: %+v, err: %v", members, err)
	}
}

func TestService_UpdateMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/chats/updateMembers/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	err := svc.UpdateMembers(context.Background(), UpdateMembersRequest{
		ChatID:  "chat1",
		Members: []User{{Login: "u1"}},
	}.WithSendNotifications(false))
	if err != nil {
		t.Fatalf("UpdateMembers error: %v", err)
	}
}
