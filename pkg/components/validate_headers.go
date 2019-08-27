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

func incomingMatchesAllowed(allowed map[string][]string, incoming map[string][]string) (bool, error) {
	allowedValuesFound := false
	for allowedKey, allowedValues := range allowed {
		matchedIncomingHeaderValues := incoming[http.CanonicalHeaderKey(allowedKey)]
		for _, item := range allowedValues {
			allowedValuesFound = contains(matchedIncomingHeaderValues, item)
			if !allowedValuesFound {
				return allowedValuesFound, fmt.Errorf("no matching values for header %s. given values %v. acceptable values %v", allowedKey, allowedValues, matchedIncomingHeaderValues)
			}
		}
	}
	return allowedValuesFound, nil
}

func (r *validateHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resultsMatch, err := incomingMatchesAllowed(r.Allowed, req.Header)
	if err != nil || !resultsMatch {
		return newError(http.StatusBadRequest, "header validation failed"), fmt.Errorf("%s", err)
	}
	return r.Wrapped.RoundTrip(req)
}

// ValidateHeaderConfig is used to validate a map of headers and their allowed values against an incoming requests headers
type ValidateHeaderConfig struct {
	Allowed map[string][]string `description:"Map of allowed headers and their values"`
}

// Name of the config root
func (*ValidateHeaderConfig) Name() string {
	return "validateheaders"
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
