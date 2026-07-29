// Package polls provides API methods to manage polls in Yandex Messenger.
package polls

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Requester defines the interface for HTTP client used by the service.
type Requester interface {
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
}

// Service provides methods to interact with polls API.
type Service struct {
	client Requester
}

// NewService creates a new instance of Service.
func NewService(client Requester) *Service {
	return &Service{client: client}
}

// CreateRequest represents parameters for creating a poll.
type CreateRequest struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	Title               string                `json:"title"`
	Answers             []string              `json:"answers"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
}

// Create sends a poll to a chat or user.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if len(req.Answers) == 0 {
		return nil, errors.New("at least one answer is required")
	}

	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/createPoll/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePoll is an alias for Create to provide consistent DX with CreatePollRequest.
func (s *Service) CreatePoll(ctx context.Context, req CreateRequest) (*types.SendResponse, error) {
	return s.Create(ctx, req)
}

// GetResultsRequest represents parameters for fetching poll results.
type GetResultsRequest struct {
	ChatID     types.ChatID
	Login      types.UserLogin
	MessageID  types.MessageID
	InviteHash string
	ThreadID   types.ThreadID
}

// GetResultsResponse represents the response containing poll results.
type GetResultsResponse struct {
	Ok         bool           `json:"ok"`
	VotedCount int            `json:"voted_count"`
	Answers    map[string]int `json:"answers"`
}

// GetResults retrieves the voting statistics for a poll by its message ID.
func (s *Service) GetResults(ctx context.Context, req GetResultsRequest) (*GetResultsResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.MessageID == 0 {
		return nil, errors.New("message_id is required")
	}

	q := url.Values{}
	if req.ChatID != "" {
		q.Set("chat_id", string(req.ChatID))
	}
	if req.Login != "" {
		q.Set("login", string(req.Login))
	}
	q.Set("message_id", strconv.FormatInt(int64(req.MessageID), 10))

	if req.InviteHash != "" {
		q.Set("invite_hash", req.InviteHash)
	}
	if req.ThreadID != 0 {
		q.Set("thread_id", strconv.FormatInt(int64(req.ThreadID), 10))
	}

	u := url.URL{Path: "polls/getResults/", RawQuery: q.Encode()}

	var resp GetResultsResponse
	if err := s.client.MakeRequest(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVotersRequest represents parameters for fetching a list of users who voted for a specific answer.
//
// Builder methods available:
//   - WithLimit(limit int): Sets the maximum number of voters to return (pointer field to allow 0).
type GetVotersRequest struct {
	ChatID     types.ChatID    // Optional if Login is provided
	Login      types.UserLogin // Optional if ChatID is provided
	MessageID  types.MessageID // Required
	AnswerID   *int            // Required (answer_id)
	Limit      *int            // Optional
	Cursor     string          // Optional (start with "0" or leave empty)
	InviteHash string          // Optional
	ThreadID   types.ThreadID  // Optional
}

// WithLimit sets the maximum number of voters to return.
func (r GetVotersRequest) WithLimit(limit int) GetVotersRequest {
	r.Limit = &limit
	return r
}

// PollVoter represents a user who voted in a poll and the time of the vote.
type PollVoter struct {
	Timestamp int64        `json:"timestamp"`
	User      types.Sender `json:"user"`
}

// GetVotersResponse represents the response containing the voters list.
type GetVotersResponse struct {
	Ok         bool            `json:"ok"`
	AnswerID   int             `json:"answer_id,omitempty"`
	VotedCount int             `json:"voted_count,omitempty"`
	Cursor     json.RawMessage `json:"cursor,omitempty"`
	Votes      []PollVoter     `json:"votes"`
}

// NextCursor returns the cursor string for the next page of voters.
// It returns an empty string if there are no more voters (e.g. when the API returns {}).
func (r *GetVotersResponse) NextCursor() string {
	c := string(r.Cursor)
	if c == "{}" || c == `""` || c == "" || c == "null" {
		return ""
	}
	// The cursor might be returned as a JSON string like `"my_cursor"`, so we trim the quotes
	if len(c) >= 2 && c[0] == '"' && c[len(c)-1] == '"' {
		return c[1 : len(c)-1]
	}
	return c
}

// GetVoters retrieves the list of users who voted for a specific answer in a poll.
func (s *Service) GetVoters(ctx context.Context, req GetVotersRequest) (*GetVotersResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.MessageID == 0 {
		return nil, errors.New("message_id is required")
	}
	if req.AnswerID == nil {
		return nil, errors.New("answer_id is required")
	}

	q := url.Values{}
	if req.ChatID != "" {
		q.Set("chat_id", string(req.ChatID))
	}
	if req.Login != "" {
		q.Set("login", string(req.Login))
	}
	q.Set("message_id", strconv.FormatInt(int64(req.MessageID), 10))
	q.Set("answer_id", strconv.Itoa(*req.AnswerID))

	if req.Limit != nil {
		q.Set("limit", strconv.Itoa(*req.Limit))
	}
	if req.Cursor != "" {
		q.Set("cursor", req.Cursor)
	}
	if req.InviteHash != "" {
		q.Set("invite_hash", req.InviteHash)
	}
	if req.ThreadID != 0 {
		q.Set("thread_id", strconv.FormatInt(int64(req.ThreadID), 10))
	}

	u := url.URL{Path: "polls/getVoters/", RawQuery: q.Encode()}

	var resp GetVotersResponse
	if err := s.client.MakeRequest(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
