package types

import "strings"

// IsCommand checks if the update contains a text message that is a command (starts with "/").
func (u *Update) IsCommand() bool {
	if u == nil || strings.TrimSpace(u.Text) == "" {
		return false
	}
	parts := strings.Fields(u.Text)
	if len(parts) == 0 {
		return false
	}
	return strings.HasPrefix(parts[0], "/")
}

// Command extracts the command name without the leading slash.
// For example, if the text is "/start params", it returns "start".
// If it's not a command, it returns an empty string.
func (u *Update) Command() string {
	if u == nil || !u.IsCommand() {
		return ""
	}
	parts := strings.Fields(u.Text)
	if len(parts) == 0 {
		return ""
	}
	// Remove the "/"
	cmd := strings.TrimPrefix(parts[0], "/")
	if idx := strings.Index(cmd, "@"); idx != -1 {
		return cmd[:idx]
	}
	return cmd
}

// CommandArguments extracts the arguments passed after a command.
// For example, if the text is "/search Go tutorials", it returns "Go tutorials".
func (u *Update) CommandArguments() string {
	if u == nil || !u.IsCommand() {
		return ""
	}
	parts := strings.Fields(u.Text)
	if len(parts) > 1 {
		return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(u.Text), parts[0]))
	}
	return ""
}

// IsPrivate checks if the update came from a private chat (direct message).
func (u *Update) IsPrivate() bool {
	if u == nil {
		return false
	}
	if u.Chat == nil && u.From != nil {
		return true // Yandex quirk: Chat is nil in some private messages
	}
	if u.Chat != nil {
		if u.Chat.Type == "private" {
			return true
		}
		// Fallback for Yandex: private chats contain "_" (uuid_uuid or login_login) and do not contain "/"
		if strings.Contains(string(u.Chat.ID), "_") && !strings.Contains(string(u.Chat.ID), "/") {
			return true
		}
	}
	return false
}

// IsGroup checks if the update came from a group or channel.
func (u *Update) IsGroup() bool {
	if u == nil {
		return false
	}
	if u.Chat != nil {
		if u.Chat.Type == "group" || u.Chat.Type == "channel" {
			return true
		}
		// Fallback for Yandex: groups and channels start with "0/0/" or "1/0/" (contain "/")
		if strings.Contains(string(u.Chat.ID), "/") {
			return true
		}
	}
	return false
}
