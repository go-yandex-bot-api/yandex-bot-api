package core

import (
	"errors"
	"fmt"
	"net/http"
)

// Common API Sentinel errors for use with errors.Is.
var (
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrRateLimited  = errors.New("rate limit exceeded")
	ErrServerError  = errors.New("internal server error")
)

// APIError represents an error returned by the Yandex Messenger API.
type APIError struct {
	StatusCode  int    // HTTP status code (e.g., 400, 429, 500)
	Description string // Raw response body or parsed description from the server
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("API request error (status %d): %s", e.StatusCode, e.Description)
	}
	return fmt.Sprintf("API request error (status %d)", e.StatusCode)
}

// Is allows errors.Is to match specific HTTP status codes to sentinel errors.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrBadRequest:
		return e.StatusCode == http.StatusBadRequest
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrConflict:
		return e.StatusCode == http.StatusConflict
	case ErrRateLimited:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrServerError:
		return e.StatusCode >= 500 && e.StatusCode < 600
	}
	return false
}

// GetStatusCode returns the HTTP status code of the error.
func (e *APIError) GetStatusCode() int {
	return e.StatusCode
}
