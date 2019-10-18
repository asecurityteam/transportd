package components

import (
	"context"
	"fmt"
	"net/http"

	"github.com/asecurityteam/transport"
)

// RequestHeaderConfig configures automated header injection.
type RequestHeaderConfig struct {
	Names  []string `description:"List of header names to inject."`
	Values []string `description:"List of header values to inject."`
}

// Name of the config root.
func (*RequestHeaderConfig) Name() string {
	return "requestheaderinject"
}

// RequestHeaderComponent implements the settings.Component interface.
type RequestHeaderComponent struct{}

// RequestHeader satisfies the NewComponent signature.
func RequestHeader(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &RequestHeaderComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*RequestHeaderComponent) Settings() *RequestHeaderConfig {
	return &RequestHeaderConfig{Names: []string{}, Values: []string{}}
}

func makeRequestDecorator(name string, value string) transport.Decorator {
	return transport.NewHeader(func(*http.Request) (string, string) {
		return name, value
	})
}

// New generates the middleware.
func (*RequestHeaderComponent) New(_ context.Context, conf *RequestHeaderConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	if len(conf.Names) != len(conf.Values) {
		return nil, fmt.Errorf(
			"header mismatch. %d names. %d values. these must match",
			len(conf.Names),
			len(conf.Values),
		)
	}
	ch := make(transport.Chain, 0, len(conf.Names))
	for offset, name := range conf.Names {
		ch = append(ch, makeRequestDecorator(name, conf.Values[offset]))
	}
	return ch.Apply, nil
}

// ResponseHeaderConfig configures automated header injection.
type ResponseHeaderConfig struct {
	Names  []string `description:"List of header names to inject."`
	Values []string `description:"List of header values to inject."`
}

// Name of the config root.
func (*ResponseHeaderConfig) Name() string {
	return "responseheaderinject"
}

// ResponseHeaderComponent implements the settings.Component interface.
type ResponseHeaderComponent struct{}

// ResponseHeader satisfies the NewComponent signature.
func ResponseHeader(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &ResponseHeaderComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*ResponseHeaderComponent) Settings() *ResponseHeaderConfig {
	return &ResponseHeaderConfig{Names: []string{}, Values: []string{}}
}

func makeResponseDecorator(name string, value string) transport.Decorator {
	return transport.NewHeaders(nil, func(*http.Response) (string, string) {
		return name, value
	})
}

// New generates the middleware.
func (*ResponseHeaderComponent) New(_ context.Context, conf *ResponseHeaderConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	if len(conf.Names) != len(conf.Values) {
		return nil, fmt.Errorf(
			"header mismatch. %d names. %d values. these must match",
			len(conf.Names),
			len(conf.Values),
		)
	}
	ch := make(transport.Chain, 0, len(conf.Names))
	for offset, name := range conf.Names {
		ch = append(ch, makeResponseDecorator(name, conf.Values[offset]))
	}
	return ch.Apply, nil
}
