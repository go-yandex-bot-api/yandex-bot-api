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

func TestService_SendImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendImage/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("failed to get multipart reader: %v", err)
		}
		var foundImage, foundChatID, foundText bool
		var imageContentType string
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("failed to read next part: %v", err)
			}
			switch part.FormName() {
			case "image":
				foundImage = true
				imageContentType = part.Header.Get("Content-Type")
			case "chat_id":
				foundChatID = true
			case "text":
				foundText = true
			}
			_ = part.Close()
		}
		if !foundImage || !foundChatID || !foundText {
			t.Errorf("missing parts: image=%v, chat_id=%v, text=%v", foundImage, foundChatID, foundText)
		}
		if imageContentType != "image/png" {
			t.Errorf("expected Content-Type image/png, got %s", imageContentType)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 301,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	// Validation checks
	_, err := svc.SendImage(context.Background(), SendImageRequest{})
	if err == nil {
		t.Error("expected error for empty request, got nil")
	}

	_, err = svc.SendImage(context.Background(), SendImageRequest{ChatID: "chat1"})
	if err == nil {
		t.Error("expected error for missing FilePath/Stream, got nil")
	}

	resp, err := svc.SendImage(context.Background(), SendImageRequest{
		ChatID:   "chat1",
		Stream:   bytes.NewReader([]byte("fake png bytes")),
		FileName: "picture.png",
		Text:     "image caption",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MessageID != 301 {
		t.Errorf("expected MessageID 301, got %d", resp.MessageID)
	}
}

func TestService_SendGallery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/sendGallery/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("failed to get multipart reader: %v", err)
		}
		imageParts := 0
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("failed to read next part: %v", err)
			}
			if part.FormName() == "images" {
				imageParts++
				ct := part.Header.Get("Content-Type")
				if ct != "image/png" && ct != "image/jpeg" {
					t.Errorf("unexpected gallery image Content-Type: %s", ct)
				}
			}
			_ = part.Close()
		}
		if imageParts != 2 {
			t.Errorf("expected 2 image parts, got %d", imageParts)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 302,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	// Validation checks
	_, err := svc.SendGallery(context.Background(), SendGalleryRequest{})
	if err == nil {
		t.Error("expected error for empty request, got nil")
	}

	_, err = svc.SendGallery(context.Background(), SendGalleryRequest{ChatID: "chat1"})
	if err == nil {
		t.Error("expected error for missing FilePaths/Streams, got nil")
	}

	resp, err := svc.SendGallery(context.Background(), SendGalleryRequest{
		ChatID: "chat1",
		Streams: []io.Reader{
			bytes.NewReader([]byte("fake png")),
			bytes.NewReader([]byte("fake jpeg")),
		},
		FileNames: []string{"img1.png", "img2.jpg"},
		Text:      "gallery caption",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MessageID != 302 {
		t.Errorf("expected MessageID 302, got %d", resp.MessageID)
	}
}

func TestService_ShareImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/shareImage/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{Ok: true, MessageID: 303})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	// Validation
	_, err := svc.ShareImage(context.Background(), ShareImageRequest{})
	if err == nil {
		t.Error("expected error for empty request, got nil")
	}
	req := ShareImageRequest{ChatID: "chat1"}
	_, err = svc.ShareImage(context.Background(), req)
	if err == nil {
		t.Error("expected error for missing FileID, got nil")
	}

	req.Image.FileID = "img123"
	req.Image.Width = 800
	req.Image.Height = 600
	resp, err := svc.ShareImage(context.Background(), req)
	if err != nil || resp.MessageID != 303 {
		t.Fatalf("unexpected response %v or error %v", resp, err)
	}
}

func TestService_ShareGallery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/shareGallery/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{Ok: true, MessageID: 304})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	// Validation
	_, err := svc.ShareGallery(context.Background(), ShareGalleryRequest{})
	if err == nil {
		t.Error("expected error for empty request, got nil")
	}
	req := ShareGalleryRequest{ChatID: "chat1"}
	_, err = svc.ShareGallery(context.Background(), req)
	if err == nil {
		t.Error("expected error for empty images, got nil")
	}

	req.Images = []ShareImageItem{{FileID: "f1", Width: 100, Height: 100}}
	resp, err := svc.ShareGallery(context.Background(), req)
	if err != nil || resp.MessageID != 304 {
		t.Fatalf("unexpected response %v or error %v", resp, err)
	}
}

func TestMimeByExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"photo.png", contentTypeImagePNG},
		{"photo.PNG", contentTypeImagePNG},
		{"picture.jpg", "image/jpeg"},
		{"picture.jpeg", "image/jpeg"},
		{"anim.gif", "image/gif"},
		{"banner.webp", "image/webp"},
		{"file.unknown", contentTypeOctetStream},
		{"noext", contentTypeOctetStream},
	}
	for _, tt := range tests {
		if got := mimeByExtension(tt.filename); got != tt.want {
			t.Errorf("mimeByExtension(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}
