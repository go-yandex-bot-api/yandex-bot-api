// Package users provides API methods to manage users in Yandex Messenger.
package users

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Requester defines the interface for HTTP client used by the service.
type Requester interface {
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
}

// Service provides methods to interact with users API.
type Service struct {
	client Requester
}

// NewService creates a new instance of Service.
func NewService(client Requester) *Service {
	return &Service{client: client}
}

// GetMe fetches basic information about the bot (equivalent to 'bot-info' / 'self/get').
func (s *Service) GetMe(ctx context.Context) (*types.BotInfo, error) {
	var info types.BotInfo
	if err := s.client.MakeRequest(ctx, http.MethodGet, "self/get/", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetSelf is an alias for GetMe to provide intuitive DX.
func (s *Service) GetSelf(ctx context.Context) (*types.BotInfo, error) {
	return s.GetMe(ctx)
}

// GetUserLinkRequest represents parameters for getting user links.
type GetUserLinkRequest struct {
	Login types.UserLogin `json:"login"`
}

// UserLinkResponse represents the response containing user links.
type UserLinkResponse struct {
	Ok       bool   `json:"ok"`
	ID       string `json:"id"`
	ChatLink string `json:"chat_link"`
	CallLink string `json:"call_link"`
}

// GetUserLink returns links to open a chat or call a user.
func (s *Service) GetUserLink(ctx context.Context, req GetUserLinkRequest) (*UserLinkResponse, error) {
	if req.Login == "" {
		return nil, errors.New("login is required")
	}
	q := url.Values{}
	q.Set("login", string(req.Login))
	u := url.URL{Path: "users/getUserLink/", RawQuery: q.Encode()}

	var resp UserLinkResponse
	if err := s.client.MakeRequest(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
