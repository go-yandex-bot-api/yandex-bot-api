// Package files provides API methods for file and media operations.
package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

//nolint:revive // exported function with intentional parameter naming for API clarity
type Requester interface {
	MakeStreamRequest(ctx context.Context, method, endpoint string, payload interface{}) (io.ReadCloser, error)
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
	MakeMultipartRequest(ctx context.Context, endpoint string, mc types.MultipartPayload, dest interface{}) error
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type Service struct {
	client Requester
}

//nolint:revive // exported function with intentional parameter naming for API clarity
func NewService(client Requester) *Service {
	return &Service{client: client}
}

// GetFileRequest represents parameters for fetching a file.

//nolint:revive // exported function with intentional parameter naming for API clarity
type GetFileRequest struct {
	FileID types.FileID `json:"file_id"`
}

// GetFile retrieves a file from the server by its file_id.
// It returns an io.ReadCloser to stream the file content.
// The caller is responsible for closing it!

//nolint:revive // exported function with intentional parameter naming for API clarity
func (s *Service) GetFile(ctx context.Context, req GetFileRequest) (io.ReadCloser, error) {
	if req.FileID == "" {
		return nil, errors.New("file_id is required")
	}
	q := url.Values{}
	q.Set("file_id", string(req.FileID))
	u := url.URL{Path: "messages/getFile/", RawQuery: q.Encode()}
	return s.client.MakeStreamRequest(ctx, http.MethodGet, u.String(), nil)
}

// GetFileByID retrieves a file directly by its FileID without creating a GetFileRequest struct.
func (s *Service) GetFileByID(ctx context.Context, fileID types.FileID) (io.ReadCloser, error) {
	return s.GetFile(ctx, GetFileRequest{FileID: fileID})
}

type singleFilePayload struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	Text                string                `json:"text,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
	FilePath            string                `json:"-"`
	Stream              io.Reader             `json:"-"`
	FileName            string                `json:"-"`
	FieldName           string                `json:"-"`
	Endpoint            string                `json:"-"`
}

func (p singleFilePayload) Method() string { return p.Endpoint }

func (p singleFilePayload) Payload() interface{} { return p }

func (p singleFilePayload) Files() []types.RequestFile {
	fileName := p.FileName
	if fileName == "" && p.FilePath != "" {
		fileName = filepath.Base(p.FilePath)
	}
	return []types.RequestFile{{FieldName: p.FieldName, FileName: fileName, FilePath: p.FilePath, Stream: p.Stream}}
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type SendFileRequest struct {
	ChatID              types.ChatID
	Login               types.UserLogin
	FilePath            string
	Stream              io.Reader
	FileName            string
	Text                string
	Important           bool
	MessageID           types.MessageID
	ReplyMessageID      types.MessageID
	ThreadID            types.ThreadID
	DisableNotification bool
	SuggestButtons      *types.SuggestButtons
	ActionButtons       *types.ActionButtons
}

// SendFile sends a file to a chat.
//
//nolint:dupl // similar upload functions for different media types intentionally share structure
func (s *Service) SendFile(ctx context.Context, req SendFileRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.FilePath == "" && req.Stream == nil {
		return nil, errors.New("either FilePath or Stream must be provided")
	}
	payload := singleFilePayload{
		ChatID:              req.ChatID,
		Login:               req.Login,
		FilePath:            req.FilePath,
		Stream:              req.Stream,
		FileName:            req.FileName,
		Text:                req.Text,
		Important:           req.Important,
		MessageID:           req.MessageID,
		ReplyMessageID:      req.ReplyMessageID,
		ThreadID:            req.ThreadID,
		DisableNotification: req.DisableNotification,
		SuggestButtons:      req.SuggestButtons,
		ActionButtons:       req.ActionButtons,
		FieldName:           "document",
		Endpoint:            "messages/sendFile/",
	}
	var resp types.SendResponse
	if err := s.client.MakeMultipartRequest(ctx, payload.Method(), payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type ShareFileRequest struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
	Text                string                `json:"text,omitempty"`
	Filename            string                `json:"filename,omitempty"`
	Document            struct {
		FileID types.FileID `json:"file_id"`
	} `json:"document"`
}

//nolint:revive // exported function with intentional parameter naming for API clarity
func (s *Service) ShareFile(ctx context.Context, req ShareFileRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.Document.FileID == "" {
		return nil, errors.New("document.file_id is required")
	}
	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/shareFile/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type SendImageRequest struct {
	ChatID              types.ChatID
	Login               types.UserLogin
	FilePath            string
	Stream              io.Reader
	FileName            string
	Text                string
	Important           bool
	MessageID           types.MessageID
	ReplyMessageID      types.MessageID
	ThreadID            types.ThreadID
	DisableNotification bool
	SuggestButtons      *types.SuggestButtons
	ActionButtons       *types.ActionButtons
}

// SendImage sends an image to a chat.
//
//nolint:dupl // similar upload functions for different media types intentionally share structure
func (s *Service) SendImage(ctx context.Context, req SendImageRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.FilePath == "" && req.Stream == nil {
		return nil, errors.New("either FilePath or Stream must be provided")
	}
	payload := singleFilePayload{
		ChatID:              req.ChatID,
		Login:               req.Login,
		FilePath:            req.FilePath,
		Stream:              req.Stream,
		FileName:            req.FileName,
		Text:                req.Text,
		Important:           req.Important,
		MessageID:           req.MessageID,
		ReplyMessageID:      req.ReplyMessageID,
		ThreadID:            req.ThreadID,
		DisableNotification: req.DisableNotification,
		SuggestButtons:      req.SuggestButtons,
		ActionButtons:       req.ActionButtons,
		FieldName:           "image",
		Endpoint:            "messages/sendImage/",
	}
	var resp types.SendResponse
	if err := s.client.MakeMultipartRequest(ctx, payload.Method(), payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type ShareImageRequest struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
	Text                string                `json:"text,omitempty"`
	Image               struct {
		FileID types.FileID `json:"file_id"`
		Width  int          `json:"width,omitempty"`
		Height int          `json:"height,omitempty"`
	} `json:"image"`
}

//nolint:revive // exported function with intentional parameter naming for API clarity
func (s *Service) ShareImage(ctx context.Context, req ShareImageRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if req.Image.FileID == "" {
		return nil, errors.New("image.file_id is required")
	}
	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/shareImage/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type sendGalleryPayload struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	Text                string                `json:"text,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
	FilePaths           []string              `json:"-"`
	Streams             []io.Reader           `json:"-"`
	FileNames           []string              `json:"-"`
}

func (p sendGalleryPayload) Method() string { return "messages/sendGallery/" }

func (p sendGalleryPayload) Payload() interface{} { return p }

func (p sendGalleryPayload) Files() []types.RequestFile {
	files := make([]types.RequestFile, 0, len(p.FilePaths)+len(p.Streams))
	for _, path := range p.FilePaths {
		files = append(files, types.RequestFile{FieldName: "images", FileName: filepath.Base(path), FilePath: path})
	}
	for i, stream := range p.Streams {
		fileName := fmt.Sprintf("image%d", i)
		if i < len(p.FileNames) && p.FileNames[i] != "" {
			fileName = p.FileNames[i]
		}
		files = append(files, types.RequestFile{FieldName: "images", FileName: fileName, Stream: stream})
	}
	return files
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type SendGalleryRequest struct {
	ChatID              types.ChatID
	Login               types.UserLogin
	FilePaths           []string
	Streams             []io.Reader
	FileNames           []string
	Text                string
	Important           bool
	MessageID           types.MessageID
	ReplyMessageID      types.MessageID
	ThreadID            types.ThreadID
	DisableNotification bool
	SuggestButtons      *types.SuggestButtons
	ActionButtons       *types.ActionButtons
}

//nolint:revive // exported function with intentional parameter naming for API clarity
func (s *Service) SendGallery(ctx context.Context, req SendGalleryRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if len(req.FilePaths) == 0 && len(req.Streams) == 0 {
		return nil, errors.New("at least one FilePath or Stream must be provided")
	}
	payload := sendGalleryPayload{
		ChatID:              req.ChatID,
		Login:               req.Login,
		FilePaths:           req.FilePaths,
		Streams:             req.Streams,
		FileNames:           req.FileNames,
		Text:                req.Text,
		Important:           req.Important,
		MessageID:           req.MessageID,
		ReplyMessageID:      req.ReplyMessageID,
		ThreadID:            req.ThreadID,
		DisableNotification: req.DisableNotification,
		SuggestButtons:      req.SuggestButtons,
		ActionButtons:       req.ActionButtons,
	}
	var resp types.SendResponse
	if err := s.client.MakeMultipartRequest(ctx, payload.Method(), payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type ShareImageItem struct {
	FileID types.FileID `json:"file_id"`
	Width  int          `json:"width,omitempty"`
	Height int          `json:"height,omitempty"`
}

//nolint:revive // exported function with intentional parameter naming for API clarity
type ShareGalleryRequest struct {
	ChatID              types.ChatID          `json:"chat_id,omitempty"`
	Login               types.UserLogin       `json:"login,omitempty"`
	Text                string                `json:"text,omitempty"`
	MessageID           types.MessageID       `json:"message_id,omitempty"`
	ReplyMessageID      types.MessageID       `json:"reply_message_id,omitempty"`
	ThreadID            types.ThreadID        `json:"thread_id,omitempty"`
	DisableNotification bool                  `json:"disable_notification,omitempty"`
	Important           bool                  `json:"important,omitempty"`
	SuggestButtons      *types.SuggestButtons `json:"suggest_buttons,omitempty"`
	ActionButtons       *types.ActionButtons  `json:"action_buttons,omitempty"`
	Images              []ShareImageItem      `json:"images"`
}

//nolint:revive // exported function with intentional parameter naming for API clarity
func (s *Service) ShareGallery(ctx context.Context, req ShareGalleryRequest) (*types.SendResponse, error) {
	if req.ChatID == "" && req.Login == "" {
		return nil, errors.New("chat_id or login is required")
	}
	if len(req.Images) == 0 {
		return nil, errors.New("at least one image is required")
	}
	var resp types.SendResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/shareGallery/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
