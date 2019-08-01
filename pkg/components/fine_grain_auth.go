package components

import (
	"context"
	"fmt"
	"net/http"
)

type authHeader struct {
	Wrapped http.RoundTripper
}

func (r *authHeader) RoundTrip(req *http.Request) (*http.Response, error){
	header := req.Header
	header.Del("Authorization")
	fmt.Println(header)
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