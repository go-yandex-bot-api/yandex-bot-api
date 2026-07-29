package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

type basicResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

// MakeRequest executes an HTTP request to the Yandex API and parses the response.
//
//nolint:gocyclo // Retry logic and backoff require some complexity
func (c *Client) MakeRequest(
	ctx context.Context, method, endpoint string, payload interface{}, dest interface{},
) error {
	var jsonData []byte
	var err error

	if payload != nil {
		jsonData, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	baseURL, err := url.Parse(c.apiURL)
	if err != nil {
		return fmt.Errorf("invalid base url: %w", err)
	}
	relURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint url: %w", err)
	}
	reqURL := baseURL.ResolveReference(relURL).String()
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(jsonData)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		req.Header.Set("Authorization", "OAuth "+c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = err
			if attempt < c.maxRetries {
				timer := time.NewTimer(jitterBackoff(attempt))
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", readErr)
			if attempt < c.maxRetries {
				timer := time.NewTimer(jitterBackoff(attempt))
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
				continue
			}
			return lastErr
		}

		// Handle retryable status codes (e.g., 429 Too Many Requests, 5xx Server Errors, or custom config)
		if c.isRetryableStatus(resp.StatusCode) {
			lastErr = &APIError{StatusCode: resp.StatusCode, Description: string(respBody)}
			if attempt < c.maxRetries {
				defaultDelay := jitterBackoff(attempt)
				retryAfter := parseRetryAfter(resp.Header, defaultDelay)
				if c.debug {
					log.Printf("[YABOTAPI DEBUG] Request failed with retryable status %d. "+
						"Retrying after %s...\n", resp.StatusCode, retryAfter)
				}
				timer := time.NewTimer(retryAfter)
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			}
			continue
		}

		// Handle other non-2xx HTTP status codes
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var basic basicResponse
			if err := json.Unmarshal(respBody, &basic); err == nil && basic.Description != "" {
				return &APIError{StatusCode: resp.StatusCode, Description: basic.Description}
			}
			return &APIError{StatusCode: resp.StatusCode, Description: string(respBody)}
		}

		if len(bytes.TrimSpace(respBody)) == 0 {
			// Some endpoints (like sendTyping) return 200 OK with an empty body
			return nil
		}

		var basic basicResponse
		if err := json.Unmarshal(respBody, &basic); err != nil {
			return fmt.Errorf("failed to parse basic response structure: %w", err)
		}

		if !basic.Ok && basic.Description != "" {
			return &APIError{StatusCode: resp.StatusCode, Description: basic.Description}
		}

		if dest != nil {
			if err := json.Unmarshal(respBody, dest); err != nil {
				return fmt.Errorf("failed to decode response into dest: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// MakeMultipartRequest builds and executes a multipart/form-data request with streaming.
//
//nolint:gocyclo // complex method by nature
func (c *Client) MakeMultipartRequest(
	ctx context.Context, endpoint string, mc types.MultipartPayload, dest interface{},
) error {
	var lastErr error
	var wg sync.WaitGroup

	defer func() {
		for _, f := range mc.Files() {
			if f.Stream != nil {
				if cl, ok := f.Stream.(io.Closer); ok {
					_ = cl.Close()
				}
			}
		}
	}()
	defer wg.Wait()

	baseURL, err := url.Parse(c.apiURL)
	if err != nil {
		return fmt.Errorf("invalid base url: %w", err)
	}
	relURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint url: %w", err)
	}
	reqURL := baseURL.ResolveReference(relURL).String()

	payload := mc.Payload()
	var payloadMap map[string]interface{}

	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload for multipart: %w", err)
		}
		decoder := json.NewDecoder(bytes.NewReader(payloadBytes))
		decoder.UseNumber()
		if err := decoder.Decode(&payloadMap); err != nil {
			return fmt.Errorf("failed to unmarshal payload to map: %w", err)
		}
	}

	localMaxRetries := c.maxRetries
	for _, f := range mc.Files() {
		if f.Stream != nil {
			localMaxRetries = 0 // Disable retries if streams are used to prevent exhaustion
			break
		}
	}

	for attempt := 0; attempt <= localMaxRetries; attempt++ {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { _ = pw.Close() }()
			defer func() { _ = writer.Close() }()

			var openFiles []io.Closer
			defer func() {
				for _, f := range openFiles {
					_ = f.Close()
				}
			}()

			for key, val := range payloadMap {
				if val == nil {
					continue
				}
				switch v := val.(type) {
				case map[string]interface{}, []interface{}:
					b, err := json.Marshal(v)
					if err != nil {
						pw.CloseWithError(fmt.Errorf("failed to marshal field %s: %w", key, err))
						return
					}
					if err := writer.WriteField(key, string(b)); err != nil {
						pw.CloseWithError(fmt.Errorf("failed to write field %s: %w", key, err))
						return
					}
				default:
					// Convert float64 numbers (from UseNumber) back properly
					strVal := fmt.Sprintf("%v", val)
					if err := writer.WriteField(key, strVal); err != nil {
						pw.CloseWithError(fmt.Errorf("failed to write field %s: %w", key, err))
						return
					}
				}
			}

			// Add files using io.Copy (streaming, prevents OOM)
			for _, reqFile := range mc.Files() {
				var src io.Reader
				var fileCloser io.Closer

				if reqFile.Stream != nil {
					src = reqFile.Stream
				} else {
					f, err := os.Open(reqFile.FilePath)
					if err != nil {
						pw.CloseWithError(fmt.Errorf("failed to open file %s: %w", reqFile.FilePath, err))
						return
					}
					src = f
					fileCloser = f
				}

				if fileCloser != nil {
					openFiles = append(openFiles, fileCloser)
				}

				part, err := writer.CreateFormFile(reqFile.FieldName, filepath.Base(reqFile.FileName))
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to create form file: %w", err))
					return
				}

				if _, err := io.Copy(part, src); err != nil {
					pw.CloseWithError(fmt.Errorf("failed to copy file content: %w", err))
					return
				}
			}
		}()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, pr)
		if err != nil {
			_ = pr.Close()
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "OAuth "+c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			_ = pr.CloseWithError(err) // Unblock the writer goroutine before retrying
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = err

			if strings.Contains(err.Error(), "failed to open file") ||
				strings.Contains(err.Error(), "no such file or directory") {
				return err
			}

			if attempt < localMaxRetries {
				timer := time.NewTimer(jitterBackoff(attempt))
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", readErr)
			if attempt < localMaxRetries {
				timer := time.NewTimer(jitterBackoff(attempt))
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
				continue
			}
			return lastErr
		}

		if c.isRetryableStatus(resp.StatusCode) {
			lastErr = &APIError{StatusCode: resp.StatusCode, Description: string(respBody)}
			if attempt < localMaxRetries {
				defaultDelay := jitterBackoff(attempt)
				retryAfter := parseRetryAfter(resp.Header, defaultDelay)
				if c.debug {
					log.Printf("[YABOTAPI DEBUG] MULTIPART Request failed with retryable status %d. "+
						"Retrying after %s...\n", resp.StatusCode, retryAfter)
				}
				timer := time.NewTimer(retryAfter)
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var basic basicResponse
			if err := json.Unmarshal(respBody, &basic); err == nil && basic.Description != "" {
				return &APIError{StatusCode: resp.StatusCode, Description: basic.Description}
			}
			return &APIError{StatusCode: resp.StatusCode, Description: string(respBody)}
		}

		if len(bytes.TrimSpace(respBody)) == 0 {
			// Some endpoints return 200 OK with an empty body
			return nil
		}

		var basic basicResponse
		if err := json.Unmarshal(respBody, &basic); err != nil {
			return fmt.Errorf("failed to parse basic response structure: %w", err)
		}

		if !basic.Ok && basic.Description != "" {
			return &APIError{StatusCode: resp.StatusCode, Description: basic.Description}
		}

		if dest != nil {
			if err := json.Unmarshal(respBody, dest); err != nil {
				return fmt.Errorf("failed to decode response into dest: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// MakeStreamRequest retrieves a stream from the server.
//
//nolint:gocyclo // complex streaming logic with retry and multipart handling
func (c *Client) MakeStreamRequest(
	ctx context.Context, method, endpoint string, payload interface{},
) (io.ReadCloser, error) {
	baseURL, err := url.Parse(c.apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	relURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint url: %w", err)
	}
	reqURL := baseURL.ResolveReference(relURL).String()

	var lastErr error

	var jsonData []byte
	if payload != nil {
		var err error
		jsonData, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(jsonData)
		}

		httpReq, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "OAuth "+c.token)
		if payload != nil {
			httpReq.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = err
			if attempt < c.maxRetries {
				timer := time.NewTimer(jitterBackoff(attempt))
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return nil, ctx.Err()
				}
			}
			continue
		}

		if c.isRetryableStatus(resp.StatusCode) {
			lastErr = &APIError{StatusCode: resp.StatusCode, Description: fmt.Sprintf("retryable error: %d", resp.StatusCode)}
			defaultDelay := jitterBackoff(attempt)
			retryAfter := parseRetryAfter(resp.Header, defaultDelay)
			_ = resp.Body.Close()
			if attempt < c.maxRetries {
				if c.debug {
					log.Printf("[YABOTAPI DEBUG] STREAM Request failed with retryable status %d. "+
						"Retrying after %s...\n", resp.StatusCode, retryAfter)
				}
				timer := time.NewTimer(retryAfter)
				select {
				case <-timer.C:
					timer.Stop()
				case <-ctx.Done():
					timer.Stop()
					return nil, ctx.Err()
				}
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			desc := fmt.Sprintf("HTTP %d", resp.StatusCode)
			if readErr == nil {
				desc = string(respBody)
			}
			return nil, &APIError{StatusCode: resp.StatusCode, Description: desc}
		}

		return resp.Body, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// parseRetryAfter parses the Retry-After header value which can be either
// a number of seconds or an HTTP-date (RFC 1123).
// If it is an HTTP-date, it calculates the duration relative to the server's Date header (if present)
// or the client's current time.
func parseRetryAfter(header http.Header, fallback time.Duration) time.Duration {
	ra := header.Get("Retry-After")
	if ra == "" {
		return fallback
	}

	// Try to parse as seconds (integer)
	if seconds, err := strconv.Atoi(ra); err == nil {
		if seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
		return fallback
	}

	// Try to parse as HTTP-date (RFC 1123, etc.)
	if t, err := http.ParseTime(ra); err == nil {
		refTime := time.Now()
		if dateHeader := header.Get("Date"); dateHeader != "" {
			if st, err := http.ParseTime(dateHeader); err == nil {
				refTime = st
			}
		}
		if dur := t.Sub(refTime); dur > 0 {
			return dur
		}
		return fallback
	}

	return fallback
}

// jitterBackoff calculates an exponential backoff duration with random jitter.
// The base delay is 500ms * 2^attempt. A jitter of up to 25% is added to prevent thundering herd.
// The maximum delay is capped at 30 seconds.
func jitterBackoff(attempt int) time.Duration {
	if attempt > 20 { //nolint:mnd // prevent shift overflow
		attempt = 20
	}
	multiplier := 1 << attempt
	base := float64(500 * multiplier)        //nolint:mnd // base backoff
	jitter := (rand.Float64() * 0.25) * base //nolint:mnd,gosec // not used for crypto

	dur := time.Duration(base+jitter) * time.Millisecond
	if dur > 30*time.Second {
		return 30 * time.Second //nolint:mnd // 30 seconds is the hard cap for exponential backoff
	}
	return dur
}
