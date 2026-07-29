package core

import (
	"net/http"
	"testing"
	"time"
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
