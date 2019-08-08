package components

import (
	"context"
	"net/http"
)

type authValidationTransport struct {
	Wrapped http.RoundTripper
	AllowedGroups []string
	LdapGroupsHeaderName string
}

func Contains(s []string, target string) bool{
	for _, c := range s{
		if target == c {
			return true
		}
	}
	return false
}

func (r *authValidationTransport) RoundTrip(req *http.Request) (*http.Response, error){
	incomingLdapGroups := req.Header[r.LdapGroupsHeaderName]
	allowedList := r.AllowedGroups
	for _, g:= range incomingLdapGroups{
		if Contains(allowedList, g){
			continue
		} else {
			return newError(http.StatusUnauthorized, "missing required LDAP group"), nil
		}
	}
	return r.Wrapped.RoundTrip(req)
}

type AuthConfig struct {
	AllowedGroups []string `description:"List of ldap groups allowed to access your service"`
	LdapGroupsHeaderName string `description:"Name of the header that contains the ldap group membership of an incoming request"`
}

func (*AuthConfig) Name() string {
	return "authconfig"
}

type AuthConfigComponent struct {}

func Auth(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &AuthConfigComponent{}, nil
}

func (*AuthConfigComponent) Settings() *AuthConfig {
	return &AuthConfig{}
}

func (*AuthConfigComponent) New(_ context.Context, conf *AuthConfig) (func(tripper http.RoundTripper) http.RoundTripper, error) {
	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &authValidationTransport{
			Wrapped: wrapped,
			AllowedGroups: conf.AllowedGroups,
			LdapGroupsHeaderName: conf.LdapGroupsHeaderName,
		}
	}, nil
}