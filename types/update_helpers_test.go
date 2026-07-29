package types_test

import (
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestUpdateHelpers(t *testing.T) {
	// 1. Nil update
	var nilUpdate *types.Update
	if nilUpdate.IsPrivate() {
		t.Errorf("expected IsPrivate to be false for nil update")
	}
	if nilUpdate.IsGroup() {
		t.Errorf("expected IsGroup to be false for nil update")
	}

	// 2. Private chat without Chat object but with From
	privNoChat := &types.Update{
		From: &types.Sender{Login: "user1"},
	}
	if !privNoChat.IsPrivate() {
		t.Errorf("expected IsPrivate to be true when Chat is nil but From is set")
	}

	// 3. Private chat with Chat.Type = "private"
	privWithType := &types.Update{
		Chat: &types.Chat{Type: "private", ID: "user1_user2"},
	}
	if !privWithType.IsPrivate() {
		t.Errorf("expected IsPrivate to be true for Chat.Type = private")
	}

	// 4. Private chat with empty Chat.Type but Yandex ID format (user1_user2)
	privFallback := &types.Update{
		Chat: &types.Chat{ID: "3210ba08-ea05-925c-5363-bf3677bf5779_4a3b0208-d07c-9169-7348-05f5afeac55f"},
	}
	if !privFallback.IsPrivate() {
		t.Errorf("expected IsPrivate to be true for Yandex private chat ID format")
	}

	// 5. Group chat with Chat.Type = "group"
	groupWithType := &types.Update{
		Chat: &types.Chat{Type: "group", ID: "0/0/group_id"},
	}
	if !groupWithType.IsGroup() {
		t.Errorf("expected IsGroup to be true for Chat.Type = group")
	}

	// 6. Group chat with empty Chat.Type but Yandex ID format (0/0/group_id)
	groupFallback := &types.Update{
		Chat: &types.Chat{ID: "0/0/31d3f777-7868-4c21-9db3-a64db7cba56e"},
	}
	if !groupFallback.IsGroup() {
		t.Errorf("expected IsGroup to be true for Yandex group chat ID format")
	}
}
