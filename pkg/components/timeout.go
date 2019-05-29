package components

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultTimeoutSettingAfter = 175 * time.Millisecond
)

type timeoutRoundTripper struct {
	http.RoundTripper
	after time.Duration
}

func (m *timeoutRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var ctx, _ = context.WithTimeout(r.Context(), m.after) // nolint
	// We are intentionally skipping the usual call to `defer cancel()` here
	// because it would mark the context as canceled as soon as this method
	// returned. Normally we would want this but the http.Transport uses the
	// context from the request to manage the state of the underlying network
	// connection. Canceling the request is a signal to close the connection.
	// The closing of the connection happens asynchronously of the context
	// so quickly processed requests are fine. However, requests that take some
	// time, such as those with large response bodies that need to be copied
	// back to the caller, would regularly encounter an early termination of
	// the copy because the underlying connection is closed mid-copy.
	return m.RoundTripper.RoundTrip(r.WithContext(ctx))
}

// TimeoutConfig adjusts the timeout value for requests.
type TimeoutConfig struct {
	After time.Duration `description:"Duration after which the request is canceled."`
}

// Name of the configuration root.
func (*TimeoutConfig) Name() string {
	return "timeout"
}

// TimeoutComponent implements the settings.Component interface.
type TimeoutComponent struct{}

// Timeout satisfies the NewComponent signature.
func Timeout(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &TimeoutComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*TimeoutComponent) Settings() *TimeoutConfig {
	return &TimeoutConfig{After: defaultTimeoutSettingAfter}
}

// New generates the middleware.
func (*TimeoutComponent) New(_ context.Context, conf *TimeoutConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	return func(next http.RoundTripper) http.RoundTripper {
		return &timeoutRoundTripper{RoundTripper: next, after: conf.After}
	}, nil
}
