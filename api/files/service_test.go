package files

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestService_SendFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendFile/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 201,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.SendFile(context.Background(), SendFileRequest{
		ChatID:   "chat1",
		Stream:   bytes.NewReader([]byte("test file content")),
		FileName: "test.txt",
		Text:     "file caption",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MessageID != 201 {
		t.Errorf("expected MessageID 201, got %d", resp.MessageID)
	}
}

func TestService_ShareFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/shareFile/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 202,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	req := ShareFileRequest{
		ChatID: "chat1",
		Text:   "shared doc",
	}
	req.Document.FileID = types.FileID("file123")

	resp, err := svc.ShareFile(context.Background(), req)
	if err != nil || resp.MessageID != 202 {
		t.Fatalf("unexpected response or error: %v", err)
	}
}

func TestService_GetFileByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("binary file data"))
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	reader, err := svc.GetFileByID(context.Background(), types.FileID("f123"))
	if err != nil {
		t.Fatalf("GetFileByID error: %v", err)
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil || string(data) != "binary file data" {
		t.Fatalf("unexpected content read: %s (err: %v)", string(data), err)
	}
}
