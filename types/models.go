package types

import "encoding/json"

// ChatID is a unique identifier for the target chat.
type ChatID string

// UserLogin is a unique identifier for the target user.
type UserLogin string

// MessageID is a unique identifier for a message.
type MessageID int64

// ThreadID is a unique identifier for a message thread.
type ThreadID int64

// UpdateID is a unique identifier for an update.
type UpdateID int64

// FileID is a unique identifier for a file or image.
type FileID string

// Vote represents a vote event in a poll update.
type Vote struct {
	PollID    string   `json:"poll_id"`
	OptionIDs []string `json:"option_ids"`
}

// Poll represents a poll object inside an incoming update.
type Poll struct {
	Title       string   `json:"title"`
	Answers     []string `json:"answers"`
	MaxChoices  int      `json:"max_choices,omitempty"`
	IsAnonymous bool     `json:"is_anonymous,omitempty"`
}

// Update represents a single update from the server.
type Update struct {
	UpdateID       UpdateID    `json:"update_id"`
	MessageID      MessageID   `json:"message_id,omitempty"`
	ReplyMessageID MessageID   `json:"reply_message_id,omitempty"`
	ThreadID       ThreadID    `json:"thread_id,omitempty"`
	Timestamp      int64       `json:"timestamp,omitempty"`
	Text           string      `json:"text,omitempty"`
	From           *Sender     `json:"from,omitempty"`
	Chat           *Chat       `json:"chat,omitempty"`
	BotRequest     *BotRequest `json:"bot_request,omitempty"`
	Sticker        *Sticker    `json:"sticker,omitempty"`
	File           *File       `json:"file,omitempty"`
	Images         [][]Image   `json:"images,omitempty"`
	Vote           *Vote       `json:"vote,omitempty"`
	Poll           *Poll       `json:"poll,omitempty"`
}

// File represents a file in an update.
type File struct {
	ID   FileID `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Image represents an image in an update.
type Image struct {
	FileID FileID `json:"file_id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Size   int64  `json:"size,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Sticker represents a sticker in an update.
type Sticker struct {
	SetID string `json:"set_id"`
	ID    string `json:"id"`
}

// Sender represents the sender of the message.
type Sender struct {
	Login       UserLogin `json:"login,omitempty"`
	ID          string    `json:"id,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	Robot       bool      `json:"robot"`
}

// Chat represents the chat where the message was sent.
type Chat struct {
	Type        string `json:"type"` // "private", "group", "channel"
	ID          ChatID `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Username    string `json:"username,omitempty"`
}

// BotRequestError represents an error that occurred on the client side while executing a directive.
type BotRequestError struct {
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
}

// BotRequest represents a request from a server action (button click).
type BotRequest struct {
	ServerAction *ServerAction     `json:"server_action,omitempty"`
	ElementID    string            `json:"element_id,omitempty"`
	Errors       []BotRequestError `json:"errors,omitempty"`
}

// ServerAction represents the payload of a server action.
type ServerAction struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// BotInfo represents the information about the bot.
type BotInfo struct {
	Ok            bool      `json:"ok"`
	ID            string    `json:"id"`
	Login         UserLogin `json:"login"`
	DisplayName   string    `json:"display_name"`
	WebhookURL    string    `json:"webhook_url"`
	Organizations []int64   `json:"organizations"`
}

// SendResponse represents the response from send methods like sendText, sendImage, etc.
type SendResponse struct {
	Ok        bool      `json:"ok"`
	MessageID MessageID `json:"message_id"`
	FileID    FileID    `json:"file_id,omitempty"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	Images    []Image   `json:"images,omitempty"`
}

// GetChatID safely returns the Chat ID if present, otherwise returns an empty string.
func (u *Update) GetChatID() ChatID {
	if u == nil {
		return ""
	}
	if u.Chat != nil && u.Chat.ID != "" {
		return u.Chat.ID
	}
	if u.From != nil {
		return ChatID(u.From.Login)
	}
	return ""
}

// GetFromLogin safely returns the sender's login if present, otherwise returns an empty string.
func (u *Update) GetFromLogin() UserLogin {
	if u == nil {
		return ""
	}
	if u.From != nil {
		return u.From.Login
	}
	return ""
}
