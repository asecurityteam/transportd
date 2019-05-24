package components

import (
	"context"
	"fmt"
	"net/http"
)

type basicAuthTransport struct {
	Username string
	Password string
	Wrapped  http.RoundTripper
}

func (c *basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(c.Username, c.Password)
	return c.Wrapped.RoundTrip(r)
}

// BasicAuthConfig is used to configure HTTP basic authentication.
type BasicAuthConfig struct {
	Username string `description:"Username to use in HTTP basic authentication."`
	Password string `description:"Password to use in HTTP basic authentication."`
}

// Name of the config root.
func (c *BasicAuthConfig) Name() string {
	return "basicauth"
}

// BasicAuthComponent is an HTTP basic auth decorator plugin.
type BasicAuthComponent struct{}

// BasicAuth satisfies the NewComponent signature.
func BasicAuth(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &BasicAuthComponent{}, nil
}

// Settings generates a config populated with defaults.
func (m *BasicAuthComponent) Settings() *BasicAuthConfig {
	return &BasicAuthConfig{}
}

// New generates the middleware.
func (*BasicAuthComponent) New(ctx context.Context, conf *BasicAuthConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	if len(conf.Username) < 1 {
		return nil, fmt.Errorf("username value is empty")
	}
	if len(conf.Password) < 1 {
		return nil, fmt.Errorf("password value is empty")
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return &basicAuthTransport{
			Username: conf.Username,
			Password: conf.Password,
			Wrapped:  next,
		}
	}, nil
}
