package components

import (
	"context"
	"net/http"
)

type authHeader struct {
	Wrapped http.RoundTripper
}

func Contains(s []string, target string) bool{
	for _, c := range s{
		if target == c {
			return true
		}
	}
	return false
}

func (r *authHeader) RoundTrip(req *http.Request) (*http.Response, error){
	incomingLdapGroups := req.Header["X-Slauth-User-Groups"]
	// We can probably load this from yaml somewhere? where load this from?
	whitelist := []string{"ciso-security-all"}

	for _, g:= range incomingLdapGroups{
		if Contains(whitelist, g){
			continue
		} else {
			return newError(http.StatusUnauthorized, "missing required LDAP group"), nil
		}
	}
	return r.Wrapped.RoundTrip(req)
}

type AuthConfig struct {}

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
		return &authHeader{
			Wrapped: wrapped,
		}
	}, nil
}