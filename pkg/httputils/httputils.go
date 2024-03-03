package httputils

import (
	"net/http"
	"strings"
	"time"
)

// HTTPClientDefaultTimeout is the default timeout for HTTP clients.
const HTTPClientDefaultTimeout = 60 * time.Second

// HandlerFunc takes in a request and response writer and implements a handler.
type HandlerFunc func(w *ResponseWriter, r *http.Request)

// ServeHTTP calls the handler function.
func (f HandlerFunc) ServeHTTP(w *ResponseWriter, r *http.Request) {
	f(w, r)
}

// GetAuthorizationHeader parses HTTP authorization header.
func GetAuthorizationHeader(header http.Header, authType string) (string, bool) {
	token, ok := strings.CutPrefix(strings.TrimSpace(header.Get("Authorization")), authType)
	if !ok {
		return "", false
	}

	return strings.TrimSpace(token), true
}

// IsHTTPSuccess determines whether or not a given status code is 2xx
func IsHTTPSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// NewHTTPClient creates and returns a new HTTP client.
func NewHTTPClient(modifier func(c *http.Client)) *http.Client {
	client := &http.Client{
		Timeout: HTTPClientDefaultTimeout,
	}

	if modifier != nil {
		modifier(client)
	}

	return client
}
