package core

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockClient struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func TestLoggingClient_RedactsAuthorizationToken(t *testing.T) {
	mock := &mockClient{
		fn: func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	logger := &loggingClient{next: mock}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", http.NoBody)
	req.Header.Set("Authorization", "OAuth SecretToken123")

	resp, err := logger.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"ok":true}` {
		t.Errorf("expected response body preserved, got %s", string(body))
	}

	if req.Header.Get("Authorization") != "OAuth SecretToken123" {
		t.Errorf("original request Authorization header modified: %s", req.Header.Get("Authorization"))
	}
}

func TestLoggingClient_JSONBodyPreserved(t *testing.T) {
	mock := &mockClient{
		fn: func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			if !strings.Contains(string(body), `"hello":"world"`) {
				t.Errorf("expected request body in next client, got %s", string(body))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}, nil
		},
	}

	logger := &loggingClient{next: mock}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/test", bytes.NewBufferString(`{"hello":"world"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := logger.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()
}
