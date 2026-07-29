// Package messages provides API methods to manage messages in Yandex Messenger.
package messages

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Requester - interface for HTTP client.
type Requester interface {
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
	MakeMultipartRequest(ctx context.Context, endpoint string, mc types.MultipartPayload, dest interface{}) error
}

// Service provides methods to interact with messages API.
type Service struct {
	client Requester
}

// NewService creates a new instance of Service.
func NewService(client Requester) *Service {
	return &Service{client: client}
}

// SendTextRequest contains information about a sendText request.
type SendTextRequest struct {
	ChatID                types.ChatID          `json:"chat_id,omitempty"`
	Login                 types.UserLogin       `json:"login,omitempty"`
	Text                  string                `json:"text"`
	MessageID             types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID        types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID              types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification   bool                  `json:"disable_notification,omitempty"`
	Important             bool                  `json:"important,omitempty"`
	DisableWebPagePreview bool                  `json:"disable_web_page_preview,omitempty"`
	PayloadID             string                `json:"payload_id,omitempty"`
	SuggestButtons        *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons         *types.ActionButtons  `json:"action_buttons,omitempty"`
}

// SendText sends a text message to a chat or user.
func (s *Service) SendText(ctx context.Context, req SendTextRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.Text == "" {
		return nil, errors.New("text is required")
	}
	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/sendText/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendSystemMessageRequest contains information about a sendSystemMessage request.
type SendSystemMessageRequest struct {
	ChatID types.ChatID    `json:"chat_id,omitempty"`
	Login  types.UserLogin `json:"login,omitempty"`
	Text   string          `json:"text"`
}

// SendSystemMessage sends a system message.
func (s *Service) SendSystemMessage(ctx context.Context, req SendSystemMessageRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.Text == "" {
		return nil, errors.New("text is required")
	}
	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/sendSystemMessage/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendStickerRequest contains information about a sendSticker request.
type SendStickerRequest struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	StickerSetID        string                `json:"sticker_set_id"`
	StickerID           string                `json:"sticker_id"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
}

// SendSticker sends a sticker to a chat or user.
func (s *Service) SendSticker(ctx context.Context, req SendStickerRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.StickerSetID == "" || req.StickerID == "" {
		return nil, errors.New("sticker_set_id and sticker_id are required")
	}
	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/sendSticker/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendTypingRequest contains information about a sendTyping request.
type SendTypingRequest struct {
	ChatID   types.ChatID    `json:"chat_id,omitempty"`
	Login    types.UserLogin `json:"login,omitempty"`
	UserID   string          `json:"user_id,omitempty"`
	ThreadID types.ThreadID  `json:"thread_id,omitempty"`
}

// SendTyping sends a typing indicator to a chat or user.
func (s *Service) SendTyping(ctx context.Context, req SendTypingRequest) error {
	if req.ChatID == "" && req.Login == "" {
		return errors.New("chat_id or login is required")
	}
	return s.client.MakeRequest(ctx, http.MethodPost, "messages/sendTyping/", req, nil)
}

// Action Messages: Delete, Pin, Unpin

// DeleteMessageRequest contains information about a delete message request.
type DeleteMessageRequest struct {
	ChatID    types.ChatID    `json:"chat_id,omitempty"`
	Login     types.UserLogin `json:"login,omitempty"`
	MessageID types.MessageID `json:"message_id"`
}

// Delete deletes a message in a chat or DM.
func (s *Service) Delete(ctx context.Context, req DeleteMessageRequest) error {
	if req.ChatID == "" && req.Login == "" {
		return errors.New("chat_id or login is required")
	}
	if req.MessageID == 0 {
		return errors.New("message_id is required")
	}
	return s.client.MakeRequest(ctx, http.MethodPost, "messages/delete/", req, nil)
}

// PinMessageRequest contains information about a pin message request.
type PinMessageRequest struct {
	ChatID    types.ChatID    `json:"chat_id,omitempty"`
	Login     types.UserLogin `json:"login,omitempty"`
	MessageID types.MessageID `json:"message_id"`
}

// Pin pins a message in a chat or DM.
func (s *Service) Pin(ctx context.Context, req PinMessageRequest) error {
	if req.ChatID == "" && req.Login == "" {
		return errors.New("chat_id or login is required")
	}
	if req.MessageID == 0 {
		return errors.New("message_id is required")
	}
	return s.client.MakeRequest(ctx, http.MethodPost, "messages/pin/", req, nil)
}

// UnpinMessageRequest contains information about an unpin message request.
type UnpinMessageRequest struct {
	ChatID    types.ChatID    `json:"chat_id,omitempty"`
	Login     types.UserLogin `json:"login,omitempty"`
	MessageID types.MessageID `json:"message_id"`
}

// Unpin unpins a message in a chat or DM.
func (s *Service) Unpin(ctx context.Context, req UnpinMessageRequest) error {
	if req.ChatID == "" && req.Login == "" {
		return errors.New("chat_id or login is required")
	}
	if req.MessageID == 0 {
		return errors.New("message_id is required")
	}
	return s.client.MakeRequest(ctx, http.MethodPost, "messages/unpin/", req, nil)
}
