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
	var ctx, cancel = context.WithTimeout(r.Context(), m.after)
	defer cancel()
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
