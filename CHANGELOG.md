# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-07-29

### Added
- Complete Yandex Messenger API support (Messages, Chats, Files, Polls, Users, Updates, Webhooks).
- Built-in Router for handling commands, text, buttons, and files cleanly without heavy switch-case blocks.
- Finite State Machine (FSM) support with `MemoryStorage` for multi-step dialogues.
- `KeyboardBuilder` for easy construction of inline keyboards.
- Sub-packages `pkg/format` for text formatting and `pkg/pagination` for list pagination.
- Full unit test coverage for services and HTTP core.
- Both Long-Polling and Webhook mechanisms.
- Comprehensive `examples/` directory demonstrating project structure and use-cases.
- CI pipeline with `golangci-lint` (v2.1) and tests enabled.
