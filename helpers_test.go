package yabotapi

import (
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestNewReply(t *testing.T) {
	t.Run("nil update", func(t *testing.T) {
		req := NewReply(nil, "hello")
		if req.Text != "" {
			t.Errorf("expected empty request for nil update, got %+v", req)
		}
	})

	t.Run("reply to chat", func(t *testing.T) {
		u := &types.Update{
			Chat:      &types.Chat{ID: "group_123"},
			MessageID: 55,
			ThreadID:  10,
		}
		req := NewReply(u, "reply text")
		if req.ChatID != "group_123" || req.ReplyMessageID != 55 || req.ThreadID != 10 || req.Text != "reply text" {
			t.Errorf("unexpected reply request: %+v", req)
		}
	})

	t.Run("reply to private message without chat", func(t *testing.T) {
		u := &types.Update{
			From:      &types.Sender{Login: "john_doe"},
			MessageID: 66,
		}
		req := NewReply(u, "private reply")
		if req.Login != "john_doe" || req.ReplyMessageID != 66 || req.Text != "private reply" {
			t.Errorf("expected Login john_doe, got: %+v", req)
		}
	})
}
