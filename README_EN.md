# Yandex Messenger Bot API for Go

[![CI](https://github.com/go-yandex-bot-api/yandex-bot-api/actions/workflows/ci.yml/badge.svg)](https://github.com/go-yandex-bot-api/yandex-bot-api/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-yandex-bot-api/yandex-bot-api.svg)](https://pkg.go.dev/github.com/go-yandex-bot-api/yandex-bot-api)
[![Go Version](https://img.shields.io/github/go-mod/go-version/go-yandex-bot-api/yandex-bot-api)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`yandex-bot-api` is a Go library for developing Yandex Messenger bots.

Provides convenient interaction with the Yandex Messenger API, router, pagination, and state management (FSM).

[Читать на русском](README.md)

## Features

- **Full Yandex API Coverage**: Send standard text, attach files, upload image galleries, create interactive polls, and complex inline keyboards.
- **Built-in Router**: Elegant and strict routing for incoming messages (`HandleCommand`, `HandleText`, `HandleButton`). Allows you to replace massive, unreadable `switch-case` constructs.
- **Finite State Machine (FSM)**: Convenient user session management. Create complex multi-step dialogs (such as step-by-step surveys or shopping carts) while safely storing state data directly into the user context via `Context.FSM()`.
- **Data Streaming**: Full support for the `io.Reader` interface, allowing you to upload files without saving them to the hard drive first. Perfect for instantly sending dynamically generated reports or graphics.
- **Middlewares**: Ability to inject global interceptors at the router level. Ideal for adding loggers, authorization layers, or metrics collection systems (Prometheus/Grafana).
- **Two Operating Modes**: The library natively supports both Long Polling (great for local development) and Webhooks (the industry standard for scalable Production systems).

## Installation

Ensure you have Go version 1.23 or higher installed, and run:

```bash
go get -u github.com/go-yandex-bot-api/yandex-bot-api
```

## Quick Start

Below is the code for a simple echo bot. Notice how the `router.Context` object abstracts away the boilerplate: it automatically identifies the sender and directs the response to the correct chat without requiring you to manually operate with chat IDs.

```go
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func main() {
	// 1. Initialize the HTTP bot client with your token
	bot, err := yabotapi.NewBot("YOUR_YANDEX_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Start receiving updates in the background (Long Polling)
	updates, err := bot.Updates.GetUpdatesChannel(context.Background(), yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal(err)
	}

	// 3. Create a router to handle incoming messages
	r := router.NewRouter(bot)

	// Handle a specific command (e.g., /start)
	r.HandleCommand("start", func(c *router.Context) error {
		return c.Reply("Hello! I am a modular bot for Yandex Messenger.")
	})

	// Handle any arbitrary text message
	r.HandleText(func(c *router.Context) error {
		return c.Reply("Echo: " + c.Update.Text)
	})

	// 4. Launch the blocking routing loop
	log.Println("Bot is successfully running and ready")
	r.Start(context.Background(), updates)
}
```

## Project Architecture

The codebase is designed according to clean architecture principles and is divided into strict, independent packages:

```text
github.com/go-yandex-bot-api/yandex-bot-api/
├── api/             # Packages implementing all public Yandex API methods
├── core/            # Core library: HTTP client configuration and low-level request building
├── pkg/             # Developer tools and extensions (SDK Utilities):
│   ├── fsm/         # Interfaces and In-Memory storage for Finite State Machines
│   └── router/      # The Router, Context structure, and Middleware abstractions
├── types/           # Global data structures (Update, Message, Keyboard, etc.)
├── examples/        # Reference examples for implementing various bot features
├── aliases.go       # Convenient type aliases to keep your application imports clean
└── yabotapi.go      # The main library facade and constructors
```

## Available Services

For ease of use, all API methods are grouped into domain-specific services. After initializing the bot, you can call them directly via the fields of the `bot` struct:

* `bot.Updates` — Service for receiving incoming events (Polling).
* `bot.Webhooks` — Service for setting, deleting, and processing Webhook requests from Yandex servers.
* `bot.Messages` — Primary service for sending text, Markdown-formatted text, and inline keyboards.
* `bot.Files` — Service for media handling (uploading and downloading audio, video, images).
* `bot.Polls` — Service for creating polls and voting forms.
* `bot.Info` — Service for fetching information about participants and group chat properties.
* `bot.Users` — Service for managing user profiles.

## Advanced Bot Configuration

### 1. HTTP Client Settings
When instantiating the bot, you can pass additional options using the Functional Options pattern. This allows for fine-tuning the underlying network behavior:

```go
bot, err := yabotapi.NewBot(
    "TOKEN",
    yabotapi.WithDebug(true), // Enable detailed HTTP request dumps (for debugging only!)
)
if err != nil {
    log.Fatal(err)
}
```

### 2. Operating Modes: Polling vs Webhook

The library provides both standard methods for receiving data. The choice depends on your project's architecture:

**Long Polling (Suitable for local development)**  
The bot actively sends requests to Yandex servers to check for new messages.
```go
cfg := yabotapi.UpdateConfig{
    Offset: 0,
    Limit:  100, // Batch loading: up to 100 messages per single network request
}
updates, _ := bot.Updates.GetUpdatesChannel(context.Background(), cfg)
```

**Webhook (Industry standard for Production)**  
Yandex Messenger acts proactively and sends an HTTP request to your server the moment a user writes to the bot. This saves resources and significantly reduces latency.
```go
// 1. Tell Yandex the address of your server
bot.Webhooks.Set(context.Background(), webhooks.SetRequest{
    URL: "https://api.your-domain.com/yandex/webhook",
})

// 2. Set up a standard Go HTTP handler to process incoming POST requests
http.HandleFunc("/yandex/webhook", bot.Webhooks.ListenForWebhook(updatesChan))
```

## Reference Examples
In the [`/examples`](./examples) directory, you will find ready-to-compile examples of bots designed to solve real-world tasks:
- **`07_fsm_questionnaire`** — A survey bot featuring step-by-step user state preservation.
- **`08_pagination`** — How to use inline keyboards and complex JSON payloads to create paginated menus.
- **`09_project_structure`** — An example of how to correctly organize and separate your handler code across multiple files (Dependency Injection pattern).
