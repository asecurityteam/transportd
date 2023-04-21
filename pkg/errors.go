package transportd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func newError(code int, reason string) *http.Response {
	b, _ := json.Marshal(HTTPError{
		Code:   code,
		Status: http.StatusText(code),
		Reason: reason,
	})
	return &http.Response{
		Status:     http.StatusText(code),
		StatusCode: code,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(b)),
	}
}

// ErrorToStatusCode There can be many reasons why we couldn't get a proper response from the upstream server.
// This includes timeouts, inability to connect, or the client canceling a request.
// This cannot all be captured with one status code. This method will convert golang errors to descriptive status codes.
// context.Canceled: Something likely happened on the client side that canceled the request (such as a timeout), return 499
// context.DeadlineExceeded: Likely a timeout here in the proxy, return 504
// Default: 502
func ErrorToStatusCode(err error) int {
	if errors.Is(err, context.Canceled) {
		return 499
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}
