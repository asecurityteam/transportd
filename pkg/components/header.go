package components

import (
	"context"
	"net/http"

	"github.com/asecurityteam/transport"
)

// RequestHeaderConfig configures automated header injection.
type RequestHeaderConfig struct {
	Headers map[string][]string `description:"Map of headers to inject into requests."`
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
	return &RequestHeaderConfig{Headers: make(map[string][]string)}
}

func makeRequestDecorator(name string, value string) transport.Decorator {
	return transport.NewHeader(func(*http.Request) (string, string) {
		return name, value
	})
}

// New generates the middleware.
func (*RequestHeaderComponent) New(_ context.Context, conf *RequestHeaderConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint

	length := 0
	for _, headerValues := range conf.Headers {
		length = length + len(headerValues)
	}

	ch := make(transport.Chain, 0, length)
	for headerName, headerValues := range conf.Headers {
		for _, headerValue := range headerValues {
			ch = append(ch, makeRequestDecorator(headerName, headerValue))
		}
	}
	return ch.Apply, nil
}

// ResponseHeaderConfig configures automated header injection.
type ResponseHeaderConfig struct {
	Headers map[string][]string `description:"Map of headers to inject into responses."`
}

// Name of the config root.
func (*ResponseHeaderConfig) Name() string {
	return "responseheaderinject"
}

// ResponseHeaderComponent implements the settings.Component interface.
type ResponseHeaderComponent struct{}

// ResponseHeader satisfies the NewComponent signature.
func ResponseHeader(_ context.Context, a string, b string, c string) (interface{}, error) {
	return &ResponseHeaderComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*ResponseHeaderComponent) Settings() *ResponseHeaderConfig {
	return &ResponseHeaderConfig{Headers: make(map[string][]string)}
}

func makeResponseDecorator(name string, value string) transport.Decorator {
	return transport.NewHeaders(nil, func(*http.Response) (string, string) {
		return name, value
	})
}

// New generates the middleware.
func (*ResponseHeaderComponent) New(_ context.Context, conf *ResponseHeaderConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint

	length := 0
	for _, headerValues := range conf.Headers {
		length = length + len(headerValues)
	}

	ch := make(transport.Chain, 0, length)
	for headerName, headerValues := range conf.Headers {
		for _, headerValue := range headerValues {
			ch = append(ch, makeResponseDecorator(headerName, headerValue))
		}
	}
	return ch.Apply, nil
}
