// Package webhooks provides API methods to manage webhooks for Yandex Messenger.
package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Requester defines the interface for HTTP client used by the service.
type Requester interface {
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
}

// Service provides methods to interact with webhooks API.
type Service struct {
	client Requester
	debug  bool
}

// NewService creates a new instance of Service.
func NewService(client Requester, debug bool) *Service {
	return &Service{client: client, debug: debug}
}

// SetWebhookRequest represents the request to set or remove a webhook.
type SetWebhookRequest struct {
	WebhookURL *string `json:"webhook_url"`
}

// SetWebhook sets a new webhook URL for the bot.
func (s *Service) SetWebhook(ctx context.Context, url string) error {
	var req SetWebhookRequest
	if url == "" {
		req.WebhookURL = nil // Sending null deletes the webhook
	} else {
		req.WebhookURL = &url
	}

	return s.client.MakeRequest(ctx, http.MethodPost, "self/update/", req, nil)
}

// ListenForWebhook returns a channel of updates and an http.HandlerFunc.
func (s *Service) ListenForWebhook(ctx context.Context) (<-chan types.Update, http.HandlerFunc) {
	_ = ctx                            // Context is not currently used since we removed the dangerous channel close
	ch := make(chan types.Update, 100) //nolint:mnd // default buffer size

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Updates []types.Update `json:"updates"`
		}

		defer func() { _ = r.Body.Close() }()
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) //nolint:mnd // 10MB max body size
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Webhook body read error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if s.debug {
			log.Printf("[YABOTAPI DEBUG] Webhook Request from %s | Body: %s\n", r.RemoteAddr, string(body))
		}

		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("Webhook decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for i := range payload.Updates {
			update := payload.Updates[i]
			select {
			case ch <- update:
			case <-r.Context().Done():
				return
			default:
				log.Printf("Webhook dropped update: channel full")
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}

	return ch, handler
}
