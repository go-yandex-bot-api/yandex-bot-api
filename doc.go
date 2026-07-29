// Package yabotapi provides a Go client for the Yandex Messenger Bot API.
//
// This library is designed with a service-oriented architecture, featuring a built-in router,
// FSM, and automatic retries.
// It cleanly separates low-level HTTP transport from high-level domain services.
//
// # Quick Start
//
//	bot, err := yabotapi.NewBot("YOUR_OAUTH_TOKEN")
//
//	// Send a message
//	_, err = bot.Messages.SendText(context.Background(), yabotapi.SendTextRequest{
//	    Login: "nickname",
//	    Text:  "Hello from Yandex Bot API!",
//	})
//
// The API is divided into domain-specific services:
//   - Messages: Send text, images, files, galleries, keyboards.
//   - Chats: Manage groups, participants, admins.
//   - Users: Retrieve user information.
//   - Updates: Short-polling mechanism for receiving events.
//   - Webhooks: Set up and handle incoming webhooks.
//   - Polls: Create and manage polls.
//   - Files: Stream and download files.
//
// # Features
//
//   - Auto-Retries: The core HTTP client automatically handles 429 Too Many Requests and 5xx errors
//     using exponential backoff or the Retry-After header.
//
//   - Polling & Webhooks: Seamlessly switch between Short-Polling and Webhooks. Both approaches return
//     a <-chan types.Update which can be fed directly into the Router.
//
//   - Router: A powerful event dispatcher that routes updates based on commands (/start), button
//     clicks (Server Actions), generic text, or fallback scenarios.
//
// - FSM (Finite State Machine): Build complex dialogue flows using in-memory or custom storage.
//
//   - File Streaming: Send and download files and images using efficient multipart streaming via
//     io.Pipe, keeping memory usage low.
//
// # Smart Helpers
//
// Functions like NewReply automatically determine whether an Update originated from a direct
// message (UserLogin) or a group chat (ChatID), ensuring your response is routed correctly.
package yabotapi
