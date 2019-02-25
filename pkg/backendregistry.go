package transportd

import (
	"context"
	"net/http"
	"strings"
)

// StaticBackendRegistry is an implementation of BackendRegistry that
// operates on a static mapping. This implementation exists, largely,
// in order to control for case insensitivity in the map.
type StaticBackendRegistry struct {
	Transports map[string]http.RoundTripper
}

// NewStaticBackendRegistry initializes a StaticBackendRegistry.
func NewStaticBackendRegistry() *StaticBackendRegistry {
	return &StaticBackendRegistry{
		Transports: make(map[string]http.RoundTripper),
	}
}

// Store a base transport for a backend.
func (r *StaticBackendRegistry) Store(_ context.Context, backend string, rt http.RoundTripper) {
	backend = strings.ToUpper(backend)
	r.Transports[backend] = rt
}

// Load the transport base for a backend. Result may be nil if unset.
func (r *StaticBackendRegistry) Load(_ context.Context, backend string) http.RoundTripper {
	backend = strings.ToUpper(backend)
	return r.Transports[backend]
}
