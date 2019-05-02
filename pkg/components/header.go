package components

import (
	"context"
	"fmt"
	"net/http"

	"github.com/asecurityteam/transport"
)

// HeaderConfig configures automated header injection.
type HeaderConfig struct {
	Names  []string `description:"List of header names to inject."`
	Values []string `description:"List of header values to inject."`
}

// Name of the config root.
func (*HeaderConfig) Name() string {
	return "headerinject"
}

// HeaderComponent implements the settings.Component interface.
type HeaderComponent struct{}

// Header satisfies the NewComponent signature.
func Header(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &HeaderComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*HeaderComponent) Settings() *HeaderConfig {
	return &HeaderConfig{Names: []string{}, Values: []string{}}
}

func makeDecorator(name string, value string) transport.Decorator {
	return transport.NewHeader(func(*http.Request) (string, string) {
		return name, value
	})
}

// New generates the middleware.
func (*HeaderComponent) New(_ context.Context, conf *HeaderConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	if len(conf.Names) != len(conf.Values) {
		return nil, fmt.Errorf(
			"header mismatch. %d names. %d values. these must match",
			len(conf.Names),
			len(conf.Values),
		)
	}
	ch := make(transport.Chain, 0, len(conf.Names))
	for offset, name := range conf.Names {
		ch = append(ch, makeDecorator(name, conf.Values[offset]))
	}
	return ch.Apply, nil
}
