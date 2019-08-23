package components

import (
	"context"
	"fmt"
	"net/http"
)

type validateHeaderTransport struct {
	Wrapped http.RoundTripper
	Allowed map[string][]string
}

func contains(s []string, target string) bool {
	for _, c := range s {
		if target == c {
			return true
		}
	}
	return false
}

func incomingMatchesAllowed(allowed map[string][]string, incoming map[string][]string) bool {
	allowedValuesFound := false
	for allowedKey, allowedValues := range allowed {
		matchedIncomingHeaderValues := incoming[allowedKey]
		for _, item := range allowedValues {
			allowedValuesFound = contains(matchedIncomingHeaderValues, item)
			if !allowedValuesFound {
				return allowedValuesFound
			}
		}
	}
	return allowedValuesFound
}

func (r *validateHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if incomingMatchesAllowed(r.Allowed, req.Header) {
		return r.Wrapped.RoundTrip(req)
	}
	return newError(http.StatusBadRequest, "missing required header value"), fmt.Errorf("missing required header value")
}

// ValidateHeaderConfig is used to configure authorization based on ldap group membership sent in a header
type ValidateHeaderConfig struct {
	Allowed map[string][]string `description:"List of allowed headers and "`
}

// Name of the config root
func (*ValidateHeaderConfig) Name() string {
	return "authheaders"
}

// ValidateHeaderConfigComponent is a plugin
type ValidateHeaderConfigComponent struct{}

// ValidateHeaders satisfies the NewComponent signature
func ValidateHeaders(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &ValidateHeaderConfigComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*ValidateHeaderConfigComponent) Settings() *ValidateHeaderConfig {
	return &ValidateHeaderConfig{}
}

// New generates the middleware.
func (*ValidateHeaderConfigComponent) New(_ context.Context, conf *ValidateHeaderConfig) (func(tripper http.RoundTripper) http.RoundTripper, error) {
	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &validateHeaderTransport{
			Wrapped: wrapped,
			Allowed: conf.Allowed,
		}
	}, nil
}
