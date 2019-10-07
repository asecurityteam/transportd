package transportd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBackend is the value given when a particular route is not part
	// of a known, named backend.
	DefaultBackend = "default"
)

// NewComponent is the signature all component plugins must implement. It
// is a constructor function that will be given the current backend and path
// for which the system is generating the component. The resulting value must
// implement the Component interface from the 'settings' project.
type NewComponent func(ctx context.Context, backend string, path string, method string) (interface{}, error)

// ClientRegistry manages a set of configured http.RoundTripper implementations that
// will be used to make requests.
type ClientRegistry interface {
	Load(ctx context.Context, path string, method string) http.RoundTripper
	Store(ctx context.Context, path string, method string, rt http.RoundTripper)
}

// Backend is an extension of http.RoundTripper that give access to relevant
// features such as the host rewrite data or the connection TTL.
type Backend interface {
	http.RoundTripper
	Host() *url.URL
	Count() int
	TTL() time.Duration
}

// BackendRegistry manages a set of base http.RoundTripper implementations that
// are composed with other tools in order to create clients.
type BackendRegistry interface {
	Load(ctx context.Context, backend string) Backend
	Store(ctx context.Context, backend string, rt Backend)
}

// HTTPError is the canonical shape for all internally generated HTTP error
// responses. All components should emit this shape, with an application/json
// content type, any time the component would return an internally crafted
// response. External developers are encouraged to embed this code in your
// project rather than importing it to avoid an unnecessary reference.
//
// The purpose of standardizing on this structure is to enable users of
// this project to add a default response schema for output validation.
// For the same reasons, we recommend that projects behind this proxy
// also use this structure for errors.
type HTTPError struct {
	// Code is the HTTP status code.
	Code int `json:"code"`
	// Status is the HTTP status string.
	Status string `json:"status"`
	// Reason is the debug data.
	Reason string `json:"reason"`
}
