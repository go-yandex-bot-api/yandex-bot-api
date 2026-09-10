# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-09-10

### Fixed
- `SendImage` and `SendGallery` now send the correct `Content-Type` (`image/png`, `image/jpeg`, etc.)
  for the `image`/`images` multipart fields instead of `application/octet-stream`.
  The Yandex Messenger API strictly validates this header and returned HTTP 415 before this fix.

### Changed
- `core/request.go`: replaced `writer.CreateFormFile` (hardcoded `application/octet-stream`) with
  `writer.CreatePart` using explicit MIME headers. MIME type is resolved via the new `mimeByFilename`
  helper (extension-based), or taken from the new `ContentType` field in `RequestFile`.
- `types/multipart.go`: added optional `ContentType string` field to `RequestFile`.
  When empty, the type is inferred automatically from the file extension.
- `api/files/service.go`: `singleFilePayload.Files()` and `sendGalleryPayload.Files()` now populate
  `ContentType` for `image`/`images` fields. `SendFile` (`document` field) retains
  `application/octet-stream` as appropriate for generic binary files.

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
