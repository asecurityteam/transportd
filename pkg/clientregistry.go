package transportd

import (
	"context"
	"net/http"
	"strings"
)

// StaticClientRegistry is an implementation of ClientRegisty that operates
// on a static mapping. This exists, largely, to protect consumers from
// case insensitivity issues.
type StaticClientRegistry struct {
	Transports map[string]map[string]http.RoundTripper
}

// NewStaticClientRegistry intializes the StaticClientRegistry.
func NewStaticClientRegistry() *StaticClientRegistry {
	return &StaticClientRegistry{
		Transports: make(map[string]map[string]http.RoundTripper),
	}
}

// Store a client for the path and method.
func (r *StaticClientRegistry) Store(_ context.Context, path string, method string, rt http.RoundTripper) {
	path = strings.ToUpper(path)
	method = strings.ToUpper(method)
	if _, ok := r.Transports[path]; !ok {
		r.Transports[path] = make(map[string]http.RoundTripper)
	}
	r.Transports[path][method] = rt
}

// Load a client for the path and method. The result may be nil if a client
// was never stored.
func (r *StaticClientRegistry) Load(_ context.Context, path string, method string) http.RoundTripper {
	path = strings.ToUpper(path)
	method = strings.ToUpper(method)
	if _, ok := r.Transports[path]; !ok {
		return nil
	}
	return r.Transports[path][method]
}
