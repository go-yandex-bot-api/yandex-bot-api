package yabotapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/api/messages"
)

func TestBot_GetMe_Success(t *testing.T) {
	// Start a local mock server to simulate Yandex API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/self/get/" {
			t.Errorf("Expected path /self/get/, got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "OAuth TEST_TOKEN" {
			t.Errorf("Expected OAuth TEST_TOKEN, got %s", auth)
		}

		response := `{
			"ok": true,
			"id": "123456",
			"login": "test_bot",
			"display_name": "Test Bot"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer mockServer.Close()

	// Configure the bot to use our fake server URL instead of the real Yandex API
	bot, newErr := NewBot("TEST_TOKEN", WithAPIURL(mockServer.URL))
	if newErr != nil {
		t.Fatalf("Failed to create bot: %v", newErr)
	}
	info, err := bot.Users.GetMe(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if info.Login != "test_bot" {
		t.Errorf("Expected login 'test_bot', got %s", info.Login)
	}
}

func TestBot_Send_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/messages/sendText/" {
			t.Errorf("Expected path /messages/sendText/, got %s", r.URL.Path)
		}

		// Decode the incoming JSON payload sent by our bot
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		if payload["chat_id"] != "chat123" {
			t.Errorf("Expected chat_id 'chat123', got %v", payload["chat_id"])
		}
		if payload["text"] != "Hello World" {
			t.Errorf("Expected text 'Hello World', got %v", payload["text"])
		}

		response := `{
			"ok": true,
			"message_id": 987654
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer mockServer.Close()

	bot, newErr := NewBot("TEST_TOKEN", WithAPIURL(mockServer.URL))
	if newErr != nil {
		t.Fatalf("Failed to create bot: %v", newErr)
	}

	req := messages.SendTextRequest{ChatID: "chat123", Text: "Hello World"}
	resp, err := bot.Messages.SendText(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.MessageID != 987654 {
		t.Errorf("Expected message_id 987654, got %d", resp.MessageID)
	}
}

func TestBot_ErrorResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := `{
			"ok": false,
			"description": "Invalid token"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized) // 401
		_, _ = w.Write([]byte(response))
	}))
	defer mockServer.Close()

	bot, newErr := NewBot("BAD_TOKEN", WithAPIURL(mockServer.URL))
	if newErr != nil {
		t.Fatalf("Failed to create bot: %v", newErr)
	}
	_, err := bot.Users.GetMe(context.Background())

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectedError := "API request error (status 401): Invalid token"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%v'", expectedError, err)
	}
}
