package yabotapi

import (
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/api/chats"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/files"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/messages"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/polls"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/updates"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/users"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/webhooks"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/fsm"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// MemoryStorage represents an in-memory storage for FSM.
type MemoryStorage = fsm.MemoryStorage

// NewMemoryStorage creates a new in-memory storage with default TTL.
func NewMemoryStorage(ttl time.Duration) *MemoryStorage {
	return fsm.NewMemoryStorage(ttl)
}

// InlineSuggestButton represents a button in an inline keyboard.
type InlineSuggestButton = types.InlineSuggestButton

// Directive represents an action executed when a button is pressed.
type Directive = types.Directive

// SuggestButtons represents a keyboard grid for quick replies or inline buttons.
type SuggestButtons = types.SuggestButtons

// ActionButtons represents an array of action buttons for a message.
type ActionButtons = types.ActionButtons

// ActionButton represents a single action button.
type ActionButton = types.ActionButton

// ActionButtonIcon represents an icon for an action button.
type ActionButtonIcon = types.ActionButtonIcon

// UpdatesConfig represents the configuration for the Long Polling updates channel.
type UpdatesConfig = updates.Config

// GetUserLinkRequest represents parameters for getting a user link.
type GetUserLinkRequest = users.GetUserLinkRequest

// GetFileRequest represents parameters for fetching a file.
type GetFileRequest = files.GetFileRequest

// GetChatsRequest represents parameters for fetching bot's chats.
type GetChatsRequest = chats.GetChatsRequest

// GetMembersRequest represents parameters for fetching chat members.
type GetMembersRequest = chats.GetMembersRequest

// UpdateMembersRequest represents parameters for updating chat members.
type UpdateMembersRequest = chats.UpdateMembersRequest

// CreateChatRequest represents parameters for creating a new chat.
type CreateChatRequest = chats.CreateChatRequest

// SendTextRequest represents parameters for sending a text message.
type SendTextRequest = messages.SendTextRequest

// SendFileRequest represents parameters for sending a file.
type SendFileRequest = files.SendFileRequest

// SendImageRequest represents parameters for sending an image.
type SendImageRequest = files.SendImageRequest

// SendGalleryRequest represents parameters for sending a gallery.
type SendGalleryRequest = files.SendGalleryRequest

// CreatePollRequest represents parameters for creating a poll.
type CreatePollRequest = polls.CreateRequest

// SendSystemMessageRequest represents parameters for sending a system message.
type SendSystemMessageRequest = messages.SendSystemMessageRequest

// SendStickerRequest represents parameters for sending a sticker.
type SendStickerRequest = messages.SendStickerRequest

// SendTypingRequest represents parameters for sending a typing event.
type SendTypingRequest = messages.SendTypingRequest

// DeleteMessageRequest represents parameters for deleting a message.
type DeleteMessageRequest = messages.DeleteMessageRequest

// PinMessageRequest represents parameters for pinning a message.
type PinMessageRequest = messages.PinMessageRequest

// UnpinMessageRequest represents parameters for unpinning a message.
type UnpinMessageRequest = messages.UnpinMessageRequest

// ShareFileRequest represents parameters for sharing a file.
type ShareFileRequest = files.ShareFileRequest

// ShareImageRequest represents parameters for sharing an image.
type ShareImageRequest = files.ShareImageRequest

// ShareGalleryRequest represents parameters for sharing a gallery.
type ShareGalleryRequest = files.ShareGalleryRequest

// ShareImageItem represents an image item in a gallery.
type ShareImageItem = files.ShareImageItem

// GetResultsRequest represents parameters for getting poll results.
type GetResultsRequest = polls.GetResultsRequest

// GetVotersRequest represents parameters for getting poll voters.
type GetVotersRequest = polls.GetVotersRequest

// ChatUser represents a user in a chat.
type ChatUser = chats.User

// CreateChatResponse represents the response when creating a chat.
type CreateChatResponse = chats.CreateChatResponse

// GetChatsResponse represents the response when getting chats.
type GetChatsResponse = chats.GetChatsResponse

// GetMembersResponse represents the response when getting chat members.
type GetMembersResponse = chats.GetMembersResponse

// ChatMember represents a member in a chat.
type ChatMember = chats.ChatMember

// GetResultsResponse represents the response when getting poll results.
type GetResultsResponse = polls.GetResultsResponse

// GetVotersResponse represents the response when getting poll voters.
type GetVotersResponse = polls.GetVotersResponse

// PollVoter represents a voter in a poll.
type PollVoter = polls.PollVoter

// UserLinkResponse represents the response when getting a user link.
type UserLinkResponse = users.UserLinkResponse

// SendResponse represents the response returned after sending a message, file, or image.
type SendResponse = types.SendResponse

// FileID represents a unique file identifier.
type FileID = types.FileID

// NewUpdateConfig creates a default UpdatesConfig.
func NewUpdateConfig(offset types.UpdateID) updates.Config {
	return updates.NewConfig(offset)
}

// NewSuggestButtons creates a simple row of inline buttons.
func NewSuggestButtons(persist bool, buttons ...types.InlineSuggestButton) *types.SuggestButtons {
	return types.NewSuggestButtons(persist, buttons...)
}

// NewSuggestButtonsGrid creates a multi-row grid of inline buttons.
func NewSuggestButtonsGrid(persist bool, rows ...[]types.InlineSuggestButton) *types.SuggestButtons {
	return types.NewSuggestButtonsGrid(persist, rows...)
}

// NewOpenURIDirective creates a directive to open a URI.
func NewOpenURIDirective(uri string) types.Directive {
	return types.NewOpenURIDirective(uri)
}

// Vote represents a vote event in a poll.
type Vote = types.Vote

// NewSendMessageDirective creates a directive to send a message on behalf of the user.
func NewSendMessageDirective(text string, payload any) types.Directive {
	return types.NewSendMessageDirective(text, payload)
}

// NewServerActionDirective creates a directive to trigger a server action (callback).
func NewServerActionDirective(name string, payload any) types.Directive {
	return types.NewServerActionDirective(name, payload)
}

// NewSetElementsStateDirective creates a directive to change the state of UI elements.
func NewSetElementsStateDirective(ids []string, state string, timeoutSeconds int) types.Directive {
	return types.NewSetElementsStateDirective(ids, state, timeoutSeconds)
}

// KeyboardBuilder assists in dynamically constructing multi-row button grids.
type KeyboardBuilder = types.KeyboardBuilder

// NewKeyboardBuilder creates a builder for constructing button grids.
func NewKeyboardBuilder(persist bool) *KeyboardBuilder {
	return types.NewKeyboardBuilder(persist)
}

// NewSimpleActionButton creates a simple button triggering a server action callback without payload.
func NewSimpleActionButton(title, actionName string) types.InlineSuggestButton {
	return types.NewSimpleActionButton(title, actionName)
}

// NewActionButton creates a button triggering a server action callback.
func NewActionButton(title, actionName string, payload any) types.InlineSuggestButton {
	return types.NewActionButton(title, actionName, payload)
}

// NewURLButton creates a button that opens a public URI.
func NewURLButton(title, uri string) types.InlineSuggestButton {
	return types.NewURLButton(title, uri)
}

// NewTextButton creates a button that sends a text message.
func NewTextButton(title, text string) types.InlineSuggestButton {
	return types.NewTextButton(title, text)
}

// File represents a file object attached to a message.
type File = types.File

// Image represents an image object attached to a message.
type Image = types.Image

// Sticker represents a sticker object.
type Sticker = types.Sticker

// ServerAction represents a server action triggered by a button press.
type ServerAction = types.ServerAction

// BotRequest represents button callback details.
type BotRequest = types.BotRequest

// BotInfo represents bot details.
type BotInfo = types.BotInfo

// Poll represents a poll object attached to a message.
type Poll = types.Poll

// SetWebhookRequest represents parameters for setting a webhook.
type SetWebhookRequest = webhooks.SetWebhookRequest

// Sender represents the sender of a message or update.
type Sender = types.Sender

// RequestFile represents a file stream or path for file upload requests.
type RequestFile = types.RequestFile
