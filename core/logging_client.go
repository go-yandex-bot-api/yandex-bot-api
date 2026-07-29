package core

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

// loggingClient is a middleware that logs all outgoing HTTP requests and incoming responses.
type loggingClient struct {
	next HTTPClient
}

// Do executes the HTTP request and logs the raw request and response.
func (l *loggingClient) Do(req *http.Request) (*http.Response, error) {
	// Dump the request (without body for files, but with body for JSON to help debugging)
	dumpReqBody := strings.HasPrefix(req.Header.Get("Content-Type"), "application/json")

	var cloneReq *http.Request
	if dumpReqBody && req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		cloneReq = req.Clone(req.Context())
		cloneReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		cloneReq = req.Clone(req.Context())
	}

	if cloneReq.Header.Get("Authorization") != "" {
		cloneReq.Header.Set("Authorization", "OAuth ***REDACTED***")
	}

	reqDump, err := httputil.DumpRequestOut(cloneReq, dumpReqBody)
	if err == nil {
		log.Printf("[YABOTAPI DEBUG] HTTP Request:\n%s\n", string(reqDump))
	} else {
		log.Printf("[YABOTAPI DEBUG] Failed to dump request: %v\n", err)
	}

	// Execute the actual request
	resp, err := l.next.Do(req)
	if err != nil {
		log.Printf("[YABOTAPI DEBUG] HTTP Request failed with error: %v\n", err)
		return resp, err
	}

	// Dump response headers
	respDump, dumpErr := httputil.DumpResponse(resp, false)
	if dumpErr == nil {
		log.Printf("[YABOTAPI DEBUG] HTTP Response:\n%s\n", string(respDump))
	} else {
		log.Printf("[YABOTAPI DEBUG] Failed to dump response: %v\n", dumpErr)
	}

	// Dump JSON response body if Content-Type is application/json for easy update debugging
	if resp != nil && resp.Body != nil && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		respBodyBytes, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr == nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))
			log.Printf("[YABOTAPI DEBUG] Response Body:\n%s\n", string(respBodyBytes))
		}
	}

	return resp, err
}
