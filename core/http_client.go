package core

import "net/http"

// HTTPClient defines the interface for making HTTP requests.
// Using an interface allows users to inject custom clients (e.g., for tracing, metrics, or mocking in tests).
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
