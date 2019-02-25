package components

import (
	"context"
	"net/http"
	"time"

	"github.com/asecurityteam/transport"
)

var (
	defaultHedgingInterval = 50 * time.Millisecond
)

// HedgingConfig adds automated retries during times of excess latency.
type HedgingConfig struct {
	Interval time.Duration `description:"Duration after which to open a new request."`
}

// Name of the config root.
func (*HedgingConfig) Name() string {
	return "hedging"
}

// HedgingComponent implements the settings.Component interface.
type HedgingComponent struct{}

// Hedging satisfies the NewComponent signature.
func Hedging(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &HedgingComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*HedgingComponent) Settings() *HedgingConfig {
	return &HedgingConfig{Interval: defaultHedgingInterval}
}

// New generates the middleware.
func (*HedgingComponent) New(_ context.Context, conf *HedgingConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	return transport.NewHedger(
		transport.NewPercentJitteredBackoffPolicy(
			transport.NewFixedBackoffPolicy(conf.Interval),
			.10,
		),
	), nil
}
