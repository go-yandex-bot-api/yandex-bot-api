package core

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name       string
		header     http.Header
		fallback   time.Duration
		wantResult time.Duration
	}{
		{
			name:       "Empty Header",
			header:     http.Header{},
			fallback:   2 * time.Second,
			wantResult: 2 * time.Second,
		},
		{
			name: "Seconds format valid",
			header: http.Header{
				"Retry-After": []string{"10"},
			},
			fallback:   2 * time.Second,
			wantResult: 10 * time.Second,
		},
		{
			name: "Seconds format invalid (negative)",
			header: http.Header{
				"Retry-After": []string{"-5"},
			},
			fallback:   2 * time.Second,
			wantResult: 2 * time.Second,
		},
		{
			name: "Seconds format invalid (zero)",
			header: http.Header{
				"Retry-After": []string{"0"},
			},
			fallback:   2 * time.Second,
			wantResult: 2 * time.Second,
		},
		{
			name: "Seconds format invalid (not a number)",
			header: http.Header{
				"Retry-After": []string{"invalid"},
			},
			fallback:   2 * time.Second,
			wantResult: 2 * time.Second,
		},
		{
			name: "HTTP-date format valid with Date header",
			header: http.Header{
				"Retry-After": []string{"Fri, 10 Jul 2026 12:00:00 GMT"},
				"Date":        []string{"Fri, 10 Jul 2026 11:59:50 GMT"},
			},
			fallback:   2 * time.Second,
			wantResult: 10 * time.Second,
		},
		{
			name: "HTTP-date format valid, but Date header is invalid (fallback to relative time.Now)",
			header: http.Header{
				"Retry-After": []string{time.Now().Add(5*time.Second).UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"},
				"Date":        []string{"invalid-date"},
			},
			fallback:   2 * time.Second,
			wantResult: 5 * time.Second,
		},
		{
			name: "HTTP-date format in the past relative to Date header",
			header: http.Header{
				"Retry-After": []string{"Fri, 10 Jul 2026 11:59:50 GMT"},
				"Date":        []string{"Fri, 10 Jul 2026 12:00:00 GMT"},
			},
			fallback:   2 * time.Second,
			wantResult: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.header, tt.fallback)
			if tt.name == "HTTP-date format valid, but Date header is invalid (fallback to relative time.Now)" {
				// Allow a small tolerance (1 second) since time.Now() is dynamic
				diff := got - tt.wantResult
				if diff < 0 {
					diff = -diff
				}
				if diff > 1*time.Second {
					t.Errorf("parseRetryAfter() = %v, want %v (with 1s tolerance)", got, tt.wantResult)
				}
			} else if got != tt.wantResult {
				t.Errorf("parseRetryAfter() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestMimeByFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"pic.png", "image/png"},
		{"pic.PNG", "image/png"},
		{"pic.jpg", "image/jpeg"},
		{"pic.jpeg", "image/jpeg"},
		{"anim.gif", "image/gif"},
		{"banner.webp", "image/webp"},
		{"video.mp4", "video/mp4"},
		{"doc.pdf", "application/pdf"},
		{"archive.zip", "application/octet-stream"},
		{"noext", "application/octet-stream"},
	}
	for _, tt := range tests {
		if got := mimeByFilename(tt.filename); got != tt.want {
			t.Errorf("mimeByFilename(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}

type dummyMultipartPayload struct {
	method  string
	payload any
	files   []types.RequestFile
}

func (d dummyMultipartPayload) Method() string              { return d.method }
func (d dummyMultipartPayload) Payload() any                { return d.payload }
func (d dummyMultipartPayload) Files() []types.RequestFile { return d.files }

func TestMakeMultipartRequest(t *testing.T) {
	const (
		testFieldImage = "image"
		testMimePNG    = "image/png"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("MultipartReader error: %v", err)
		}
		var foundButtons, foundImage bool
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("NextPart error: %v", err)
			}
			switch part.FormName() {
			case "suggest_buttons":
				foundButtons = true
				ct := part.Header.Get("Content-Type")
				if ct != "application/json; charset=utf-8" {
					t.Errorf("expected buttons Content-Type application/json; charset=utf-8, got %s", ct)
				}
			case testFieldImage:
				foundImage = true
				ct := part.Header.Get("Content-Type")
				if ct != testMimePNG {
					t.Errorf("expected image Content-Type image/png, got %s", ct)
				}
			}
			_ = part.Close()
		}
		if !foundButtons || !foundImage {
			t.Errorf("missing parts: buttons=%v, image=%v", foundButtons, foundImage)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "message_id": 999})
	}))
	defer server.Close()

	client := NewClient("test-token", WithAPIURL(server.URL+"/bot/v1/"))

	payload := dummyMultipartPayload{
		method: "messages/sendImage/",
		payload: map[string]any{
			"chat_id": "c123",
			"suggest_buttons": map[string]any{
				"buttons": []any{
					map[string]any{"text": "btn"},
				},
			},
		},
		files: []types.RequestFile{
			{
				FieldName: testFieldImage,
				FileName:  "test.png",
				Stream:    bytes.NewReader([]byte("png content")),
			},
		},
	}

	var dest basicResponse
	err := client.MakeMultipartRequest(context.Background(), "messages/sendImage/", payload, &dest)
	if err != nil {
		t.Fatalf("MakeMultipartRequest error: %v", err)
	}
	if !dest.Ok {
		t.Errorf("expected ok: true, got false")
	}
}

