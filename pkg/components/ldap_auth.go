package components

import (
	"context"
	"fmt"
	"net/http"
)

type authHeader struct {
	Wrapped http.RoundTripper
	AllowedGroups []string
}

// temp function for searching through a slice until we move the whitelist into a map
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
	// We can probably load this from a yaml file somewhere? or where should we load this from?
	// This should probably be a map? to make checking the group membership faster
	allowedList := r.AllowedGroups
	fmt.Println("## INCOMING LDAP GROUPS ARE", incomingLdapGroups)
	fmt.Println("## WHITE LIST IS: ", allowedList)
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
	AllowedGroups []string `description:List of ldap groups allowed to access your service`
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
		return &authHeader{
			Wrapped: wrapped,
			AllowedGroups: conf.AllowedGroups,
		}
	}, nil
}