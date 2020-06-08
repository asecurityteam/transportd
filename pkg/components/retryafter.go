package components

import (
	"context"
	"net/http"

	"github.com/asecurityteam/transport"
)

// RetryAfterConfig enables automated retries for status code 429 with Retry-After header honoring.
type RetryAfterConfig struct {
}

// Name of the configuration root.
func (*RetryAfterConfig) Name() string {
	return "retryafter"
}

// RetryAfterComponent implements the settings.Component interface.
type RetryAfterComponent struct{}

// RetryAfter satisfies the NewComponent signature.
func RetryAfter(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &RetryAfterComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*RetryAfterComponent) Settings() *RetryAfterConfig {
	return &RetryAfterConfig{}
}

// New generates the middleware.
func (*RetryAfterComponent) New(_ context.Context, conf *RetryAfterConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	return transport.NewRetryAfter(), nil
}
