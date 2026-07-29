// Package chats provides API methods to manage chats in Yandex Messenger.
package chats

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Requester defines the interface for HTTP client used by the service.
type Requester interface {
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
}

// Service provides methods to interact with chats API.
type Service struct {
	client Requester
}

// NewService creates a new instance of Service.
func NewService(client Requester) *Service {
	return &Service{client: client}
}

// User represents a user identifier used in requests.
type User struct {
	Login types.UserLogin `json:"login"`
}

// CreateChatRequest represents parameters for creating a new chat or channel.
type CreateChatRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Admins      []User `json:"admins,omitempty"`
	Members     []User `json:"members,omitempty"` // Use only if Channel=false
	Channel     bool   `json:"channel"`
	Subscribers []User `json:"subscribers,omitempty"` // Use only if Channel=true
	Public      bool   `json:"public"`
}

// CreateChatResponse represents the response when a chat is created.
type CreateChatResponse struct {
	Ok         bool         `json:"ok"`
	ChatID     types.ChatID `json:"chat_id"`
	InviteHash string       `json:"invite_hash,omitempty"`
	InviteLink string       `json:"invite_link,omitempty"`
}

// CreateChat creates a new chat or channel.
func (s *Service) CreateChat(ctx context.Context, req CreateChatRequest) (*CreateChatResponse, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	var resp CreateChatResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "chats/create/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetChatsRequest represents pagination parameters.
//
// Builder methods available:
//   - WithLimit(limit int): Sets the maximum number of chats to return (pointer field to allow 0).
type GetChatsRequest struct {
	Limit  *int   `json:"limit,omitempty"`
	Offset string `json:"offset,omitempty"`
}

// WithLimit sets the maximum number of chats to return.
func (r GetChatsRequest) WithLimit(limit int) GetChatsRequest {
	r.Limit = &limit
	return r
}

// GetChatsResponse represents the list of chats.
type GetChatsResponse struct {
	Ok    bool         `json:"ok"`
	Chats []types.Chat `json:"data"`
}

// GetChats returns a list of chats the bot is a member of.
func (s *Service) GetChats(ctx context.Context, req GetChatsRequest) ([]types.Chat, error) {
	u, err := url.Parse("chats/get/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	q := u.Query()
	if req.Limit != nil {
		q.Set("limit", strconv.Itoa(*req.Limit))
	}
	if req.Offset != "" {
		q.Set("offset", req.Offset)
	}
	u.RawQuery = q.Encode()

	var resp GetChatsResponse
	if err := s.client.MakeRequest(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Chats, nil
}

// ChatMember represents a member of a chat or channel.
type ChatMember struct {
	GUID  string `json:"guid"`
	Login string `json:"login"`
	Role  string `json:"role"`
	IsBot bool   `json:"is_bot"`
}

// GetMembersRequest represents parameters for fetching members of a chat.
//
// Builder methods available:
//   - WithLimit(limit int): Sets the maximum number of members to return (pointer field to allow 0).
type GetMembersRequest struct {
	ChatID types.ChatID `json:"chat_id"`
	Role   string       `json:"role,omitempty"`
	Limit  *int         `json:"limit,omitempty"`
	Offset string       `json:"offset,omitempty"`
}

// WithLimit sets the maximum number of members to return.
func (r GetMembersRequest) WithLimit(limit int) GetMembersRequest {
	r.Limit = &limit
	return r
}

// GetMembersResponse represents the list of chat members.
type GetMembersResponse struct {
	Ok      bool         `json:"ok"`
	Members []ChatMember `json:"data"`
}

// GetMembers returns a list of members in a specific chat.
func (s *Service) GetMembers(ctx context.Context, req GetMembersRequest) ([]ChatMember, error) {
	if req.ChatID == "" {
		return nil, errors.New("chat_id is required")
	}
	u, err := url.Parse("chats/getMembers/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	q := u.Query()
	q.Set("chat_id", string(req.ChatID))
	if req.Role != "" {
		q.Set("role", req.Role)
	}
	if req.Limit != nil {
		q.Set("limit", strconv.Itoa(*req.Limit))
	}
	if req.Offset != "" {
		q.Set("offset", req.Offset)
	}
	u.RawQuery = q.Encode()

	var resp GetMembersResponse
	if err := s.client.MakeRequest(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Members, nil
}

// UpdateMembersRequest represents parameters for updating members in a chat.
//
// Builder methods available:
//   - WithSendNotifications(send bool): explicitly enables or disables notifications.
//     Note: The server defaults to true. To disable notifications, you MUST use this method with false.
type UpdateMembersRequest struct {
	ChatID            types.ChatID `json:"chat_id"`
	Members           []User       `json:"members,omitempty"`
	Admins            []User       `json:"admins,omitempty"`
	Subscribers       []User       `json:"subscribers,omitempty"`
	Remove            []User       `json:"remove,omitempty"`
	SendNotifications *bool        `json:"send_members_updated_notifications,omitempty"`
}

// WithSendNotifications sets whether to send a system message to the chat about the member update.
// Use WithSendNotifications(false) to perform a silent update.
func (r UpdateMembersRequest) WithSendNotifications(send bool) UpdateMembersRequest {
	r.SendNotifications = &send
	return r
}

// UpdateMembers updates the members or administrators of a chat.
func (s *Service) UpdateMembers(ctx context.Context, req UpdateMembersRequest) error {
	if req.ChatID == "" {
		return errors.New("chat_id is required")
	}
	if err := s.client.MakeRequest(ctx, http.MethodPost, "chats/updateMembers/", req, nil); err != nil {
		return err
	}
	return nil
}
