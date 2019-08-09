package components

import (
	"context"
	"net/http"
)

type validateAuthHeaderTransport struct {
	Wrapped              http.RoundTripper
	AllowedGroups        []string
	LdapGroupsHeaderName string
}

func contains(s []string, target string) bool {
	for _, c := range s {
		if target == c {
			return true
		}
	}
	return false
}

func (r *validateAuthHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	incomingLdapGroups := req.Header[r.LdapGroupsHeaderName]
	allowedList := r.AllowedGroups
	for _, g := range incomingLdapGroups {
		if contains(allowedList, g) {
			continue
		} else {
			return newError(http.StatusUnauthorized, "missing required LDAP group"), nil
		}
	}
	return r.Wrapped.RoundTrip(req)
}

// AuthConfig is used to configure authorization based on ldap group membership sent in a header
type AuthConfig struct {
	AllowedGroups        []string `description:"List of ldap groups allowed to access your service"`
	LdapGroupsHeaderName string   `description:"Name of the header that contains the ldap group membership of an incoming request"`
}

// Name of the config root
func (*AuthConfig) Name() string {
	return "authheaders"
}

// AuthConfigComponent is a plugin
type AuthConfigComponent struct{}

// ValidateAuthHeaders satisfies the NewComponent signature
func ValidateAuthHeaders(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &AuthConfigComponent{}, nil
}

// Settings generates a config populated with defaults.
func (*AuthConfigComponent) Settings() *AuthConfig {
	return &AuthConfig{}
}

// New generates the middleware.
func (*AuthConfigComponent) New(_ context.Context, conf *AuthConfig) (func(tripper http.RoundTripper) http.RoundTripper, error) {
	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &validateAuthHeaderTransport{
			Wrapped:              wrapped,
			AllowedGroups:        conf.AllowedGroups,
			LdapGroupsHeaderName: conf.LdapGroupsHeaderName,
		}
	}, nil
}
