package components

import (
	"context"
	"net/http"
	"time"

	"github.com/asecurityteam/transport"
)

var (
	defaultRetryCodes   = []int{500, 501, 502, 503, 504, 505, 506, 507, 508, 509, 510, 511}
	defaultRetryLimit   = 3
	defaultRetryBackoff = 50 * time.Millisecond
)

// RetryConfig enables automated retires for status codes.
type RetryConfig struct {
	Codes   []int         `description:"HTTP status codes that trigger a retry."`
	Limit   int           `description:"Maximum retry attempts."`
	Backoff time.Duration `description:"Time to wait between requests."`
}

// Name of the configuration root.
func (*RetryConfig) Name() string {
	return "retry"
}

// RetryComponent implements the settings.Component interface.
type RetryComponent struct{}

// Retry satisfies the NewComponent signature.
func Retry(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &RetryComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*RetryComponent) Settings() *RetryConfig {
	return &RetryConfig{
		Codes:   defaultRetryCodes,
		Backoff: defaultRetryBackoff,
		Limit:   defaultRetryLimit,
	}
}

// New generates the middleware.
func (*RetryComponent) New(_ context.Context, conf *RetryConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	return transport.NewRetrier(
		transport.NewPercentJitteredBackoffPolicy(
			transport.NewFixedBackoffPolicy(conf.Backoff),
			.20,
		),
		transport.NewLimitedRetryPolicy(
			conf.Limit,
			transport.NewStatusCodeRetryPolicy(conf.Codes...),
		),
	), nil
}
