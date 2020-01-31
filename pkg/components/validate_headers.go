package components

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type validateHeaderTransport struct {
	Wrapped http.RoundTripper
	Allowed map[string][]string
	Split   string
}

func contains(s []string, target string, delimiter string) bool {
	for _, c := range s {
		// split on a delimiter if delimiter is not empty
		if delimiter != "" {
			for _, value := range strings.Split(c, delimiter) {
				if target == strings.TrimSpace(value) {
					return true
				}
			}
		} else if target == strings.TrimSpace(c) {
			return true
		}

	}
	return false
}

func incomingMatchesAllowed(allowed map[string][]string, incoming map[string][]string, split string) error {
	checkedHeaderAndValues := make(map[string][]string)
	for allowedHeader, allowedValues := range allowed {
		// check if incoming header values have a configured allowed header present and search through them if so
		if matchedIncomingHeaderValues, present := incoming[http.CanonicalHeaderKey(allowedHeader)]; present {

			// iterate through configured allowed values for a header
			for _, allowedValue := range allowedValues {
				// keep track of headers we have checked to return in the response if none are found
				checkedHeaderAndValues[allowedHeader] = append(checkedHeaderAndValues[allowedHeader], allowedValue)
				if contains(matchedIncomingHeaderValues, allowedValue, split) {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("no matching values for headers: %s. given matching headers and values: %s", allowed, checkedHeaderAndValues)
}

func (r *validateHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := incomingMatchesAllowed(r.Allowed, req.Header, r.Split)
	if err != nil {
		return newError(http.StatusBadRequest, fmt.Sprintf("header validation failed due to: %s", err)), nil
	}
	return r.Wrapped.RoundTrip(req)
}

// ValidateHeaderConfig is used to validate a map of headers and their allowed values against an incoming requests headers
type ValidateHeaderConfig struct {
	Allowed map[string][]string `description:"Map of headers to validate and the allowed values for those headers."`
	Split   string              `description:"The delimiter to use on the associated value"`
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
	if strings.TrimSpace(conf.Split) == "" { // revert to default
		conf.Split = ","
	} else if len(strings.TrimSpace(conf.Split)) > 1 {
		return nil, fmt.Errorf("configured split value of '%s' is not allowed", conf.Split)
	} else {
		conf.Split = strings.TrimSpace(conf.Split)
	}

	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &validateHeaderTransport{
			Wrapped: wrapped,
			Allowed: conf.Allowed,
			Split:   conf.Split,
		}
	}, nil
}
