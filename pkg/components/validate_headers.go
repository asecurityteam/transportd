package components

import (
	"context"
	"fmt"
	"net/http"
)

type validateHeaderTransport struct {
	Wrapped  http.RoundTripper
	Required map[string][]string
}

func contains(s []string, target string) bool {
	for _, c := range s {
		if target == c {
			return true
		}
	}
	return false
}

func incomingMatchesRequired(allowed map[string][]string, incoming map[string][]string) (bool, error) {
	requiredValuesFound := false
	for requiredKey, requiredValues := range allowed {
		matchedIncomingHeaderValues := incoming[http.CanonicalHeaderKey(requiredKey)]
		for _, item := range requiredValues {
			requiredValuesFound = contains(matchedIncomingHeaderValues, item)
			if !requiredValuesFound {
				return requiredValuesFound, fmt.Errorf("no matching values for header %s. given values %v. acceptable values %v", requiredKey, requiredValues, matchedIncomingHeaderValues)
			}
		}
	}
	return requiredValuesFound, nil
}

func (r *validateHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resultsMatch, err := incomingMatchesRequired(r.Required, req.Header)
	if err != nil || !resultsMatch {
		return newError(http.StatusBadRequest, "header validation failed"), fmt.Errorf("%s", err)
	}
	return r.Wrapped.RoundTrip(req)
}

// ValidateHeaderConfig is used to validate a map of headers and their required values against an incoming requests headers
type ValidateHeaderConfig struct {
	Required map[string][]string `description:"Map of headers to validate and the required values for those headers."`
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
			Wrapped:  wrapped,
			Required: conf.Required,
		}
	}, nil
}
