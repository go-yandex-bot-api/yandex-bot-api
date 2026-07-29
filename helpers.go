package yabotapi

import (
	"github.com/go-yandex-bot-api/yandex-bot-api/api/messages"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// NewReply creates a new SendTextRequest ready to be sent back to the sender of the given Update.
// It automatically detects whether the reply should go to a ChatID or a User Login.
func NewReply(u *types.Update, text string) messages.SendTextRequest {
	if u == nil {
		return messages.SendTextRequest{}
	}

	var req messages.SendTextRequest
	req.Text = text

	if u.Chat != nil && u.Chat.ID != "" {
		req.ChatID = u.Chat.ID
	} else if login := u.GetFromLogin(); login != "" {
		req.Login = login
	} else {
		return messages.SendTextRequest{}
	}
	req.ThreadID = u.ThreadID
	req.ReplyMessageID = u.MessageID

	return req
}
