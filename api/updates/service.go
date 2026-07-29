// Package updates provides API methods to receive updates from Yandex Messenger.
package updates

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

// Requester defines the interface for HTTP client used by the service.
type Requester interface {
	MakeRequest(ctx context.Context, method, endpoint string, payload interface{}, dest interface{}) error
}

// Service provides methods to interact with updates API.
type Service struct {
	client Requester
}

// NewService creates a new instance of Service.
func NewService(client Requester) *Service {
	return &Service{client: client}
}

// Config configures the update polling loop.
type Config struct {
	Limit           int
	Offset          types.UpdateID
	MinPollInterval time.Duration
	MaxPollInterval time.Duration
}

// NewConfig creates a default Config with safe default intervals.
func NewConfig(offset types.UpdateID) Config {
	return Config{
		Limit:           100, //nolint:mnd // Default and safe limit
		Offset:          offset,
		MinPollInterval: 500 * time.Millisecond, //nolint:mnd // default min interval
		MaxPollInterval: 5 * time.Second,        //nolint:mnd // default max interval
	}
}

// GetUpdatesRequest represents the internal parameters for getUpdates API.
//
// Builder methods available:
//   - WithLimit(limit int): Sets the maximum number of updates to return.
//   - WithOffset(offset types.UpdateID): Sets the offset to start fetching from (pointer to allow 0).
type GetUpdatesRequest struct {
	Limit  *int            `json:"limit,omitempty"`
	Offset *types.UpdateID `json:"offset,omitempty"`
}

// WithLimit sets the maximum number of updates to return.
func (r GetUpdatesRequest) WithLimit(limit int) GetUpdatesRequest {
	r.Limit = &limit
	return r
}

// WithOffset sets the offset to start fetching from.
func (r GetUpdatesRequest) WithOffset(offset types.UpdateID) GetUpdatesRequest {
	r.Offset = &offset
	return r
}

// GetUpdatesResponse represents the response from getUpdates API.
type GetUpdatesResponse struct {
	Ok      bool           `json:"ok"`
	Updates []types.Update `json:"updates"`
}

// GetUpdates fetches updates from the server.
func (s *Service) GetUpdates(ctx context.Context, req GetUpdatesRequest) ([]types.Update, error) {
	var resp GetUpdatesResponse
	if err := s.client.MakeRequest(ctx, http.MethodPost, "messages/getUpdates/", req, &resp); err != nil {
		return nil, err
	}
	return resp.Updates, nil
}

// GetUpdatesChannel creates a channel that receives updates.
//
//nolint:gocyclo // intentionally complex long-polling loop with retry and backoff logic
func (s *Service) GetUpdatesChannel(ctx context.Context, config Config) (<-chan types.Update, error) {
	// Validate and apply safe defaults to prevent busy-loop
	if config.MinPollInterval <= 0 {
		config.MinPollInterval = 500 * time.Millisecond //nolint:mnd // default min interval
	}
	if config.MaxPollInterval <= 0 || config.MaxPollInterval < config.MinPollInterval {
		config.MaxPollInterval = 5 * time.Second //nolint:mnd // default max interval
	}

	// First, check if a webhook is active, because getUpdates will not work.
	var info types.BotInfo
	if err := s.client.MakeRequest(ctx, http.MethodGet, "self/get/", nil, &info); err != nil {
		return nil, fmt.Errorf("failed to check bot info before polling: %w", err)
	}
	if info.WebhookURL != "" {
		return nil, fmt.Errorf("cannot start polling: webhook is active (%s)", info.WebhookURL)
	}

	ch := make(chan types.Update, 100) //nolint:mnd // default buffer size

	go func() {
		defer close(ch)
		currentInterval := config.MinPollInterval

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			req := GetUpdatesRequest{}.
				WithLimit(config.Limit).
				WithOffset(config.Offset)

			updates, err := s.GetUpdates(ctx, req)
			if err != nil {
				// Fatal errors (e.g. invalid token) should break the loop
				type statusCodeError interface {
					GetStatusCode() int
				}
				var apiErr statusCodeError
				if errors.As(err, &apiErr) && (apiErr.GetStatusCode() == 401 || apiErr.GetStatusCode() == 403) {
					return
				}

				//nolint:gosec,mnd // not used for crypto, just for jitter
				jitter := time.Duration(rand.Float64() * float64(currentInterval) * 0.25)
				timer := time.NewTimer(currentInterval + jitter)
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return
				}
				currentInterval *= 2
				if currentInterval > config.MaxPollInterval {
					currentInterval = config.MaxPollInterval
				}
				continue
			}

			if len(updates) > 0 {
				var advanced bool
				for i := range updates {
					update := updates[i]
					if update.UpdateID >= config.Offset {
						advanced = true
						config.Offset = update.UpdateID + 1
						select {
						case ch <- update:
						case <-ctx.Done():
							return
						}
					}
				}
				// If the server returned any new updates, skip backoff.
				if advanced {
					currentInterval = config.MinPollInterval
					continue
				}
			}

			//nolint:gosec,mnd // not used for crypto, just for jitter
			jitter := time.Duration(rand.Float64() * float64(currentInterval) * 0.25)
			timer := time.NewTimer(currentInterval + jitter)
			select {
			case <-timer.C:
				timer.Stop()
			case <-ctx.Done():
				timer.Stop()
				return
			}
			currentInterval *= 2
			if currentInterval > config.MaxPollInterval {
				currentInterval = config.MaxPollInterval
			}
		}
	}()

	return ch, nil
}
