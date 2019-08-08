package components

import (
	"context"
	"net/http"
	"path"
	"strings"
)

type strippingTransport struct {
	Wrapped http.RoundTripper
	Count   int
}

func (r *strippingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Path = "/" + path.Join(strings.Split(req.URL.Path, "/")[r.Count+1:]...)
	return r.Wrapped.RoundTrip(req)
}

// StripConfig contains settings for modifying the URL.
type StripConfig struct {
	Count int `description:"Number of URL segments to remove from the beginning of the path before redirect."`
}

// Name of the config root.
func (*StripConfig) Name() string {
	return "strip"
}

// StripComponent enabled modification of the URL before redirect.
type StripComponent struct{}

// Strip satisfies the NewComponent signature.
func Strip(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &StripComponent{}, nil
}

// Settings generates a config with all default values set.
func (*StripComponent) Settings() *StripConfig {
	return &StripConfig{Count: 0}
}

// New generates the middleware.
func (*StripComponent) New(_ context.Context, conf *StripConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &strippingTransport{
			Wrapped: wrapped,
			Count:   conf.Count,
		}
	}, nil
}
