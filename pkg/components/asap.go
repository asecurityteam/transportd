package components

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	asap "bitbucket.org/atlassian/go-asap"
	"github.com/SermoDigital/jose/crypto"
	"github.com/vincent-petithory/dataurl"
)

type asapValidateTransport struct {
	Wrapped   http.RoundTripper
	Validator asap.Validator
}

func (c *asapValidateTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var bearer = r.Header.Get("Authorization")
	if len(bearer) < len("Bearer ") {
		return newError(http.StatusUnauthorized, "missing bearer token"), nil
	}
	var rawToken = bearer[len("Bearer "):]
	var token, e = asap.ParseToken(rawToken)
	if e != nil {
		return newError(http.StatusUnauthorized, e.Error()), nil
	}
	e = c.Validator.Validate(token)
	if e != nil {
		return newError(http.StatusUnauthorized, e.Error()), nil
	}
	return c.Wrapped.RoundTrip(r)
}

// ASAPValidateConfig is used to configure ASAP validation.
type ASAPValidateConfig struct {
	AllowedIssuers  []string `description:"Acceptable issuer strings."`
	AllowedAudience string   `description:"Acceptable audience string."`
	KeyURLs         []string `description:"Public key download URLs."`
}

// Name of the config root.
func (m *ASAPValidateConfig) Name() string {
	return "asapvalidate"
}

// ASAPValidateComponent is an ASAP validation plugin.
type ASAPValidateComponent struct{}

// ASAPValidate satisfies the NewComponent signature.
func ASAPValidate(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &ASAPValidateComponent{}, nil
}

// Settings generates a config populated with defaults.
func (m *ASAPValidateComponent) Settings() *ASAPValidateConfig {
	return &ASAPValidateConfig{}
}

// New generates the middleware.
func (m *ASAPValidateComponent) New(ctx context.Context, conf *ASAPValidateConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	if len(conf.AllowedIssuers) < 1 {
		return nil, fmt.Errorf("allowed issuers list is empty")
	}
	if conf.AllowedAudience == "" {
		return nil, fmt.Errorf("allowed audience value is empty")
	}
	if len(conf.KeyURLs) < 1 {
		return nil, fmt.Errorf("public key url list is empty")
	}

	fetchers := make(asap.MultiKeyFetcher, 0, len(conf.KeyURLs))
	for _, keyURL := range conf.KeyURLs {
		fetchers = append(fetchers, asap.NewHTTPKeyFetcher(keyURL, &http.Client{}))
	}
	var validator = asap.NewValidatorChain(
		asap.DefaultValidator,
		asap.NewAllowedStringsValidator(asap.ClaimIssuer, conf.AllowedIssuers...),
		asap.NewAllowedAudienceValidator(conf.AllowedAudience),
		asap.NewSignatureValidator(
			asap.NewCachingFetcher(
				fetchers,
			),
		),
	)
	return func(next http.RoundTripper) http.RoundTripper {
		return &asapValidateTransport{
			Wrapped:   next,
			Validator: validator,
		}
	}, nil
}

// ASAPTokenConfig is used to configure ASAP token generation.
type ASAPTokenConfig struct {
	PrivateKey string        `description:"RSA private key to use when signing tokens."`
	KID        string        `description:"JWT kid value to include in tokens."`
	TTL        time.Duration `description:"Lifetime of a token."`
	Issuer     string        `description:"JWT issuer value to include in tokens."`
	Audiences  []string      `description:"JWT audience values to include in tokens."`
}

// Name of the config root.
func (c *ASAPTokenConfig) Name() string {
	return "asaptoken"
}

// ASAPTokenComponent is an ASAP decorator plugin.
type ASAPTokenComponent struct{}

// ASAPToken satisfies the NewComponent signature.
func ASAPToken(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &ASAPTokenComponent{}, nil
}

// NewComponent populates an ASAPToken with defaults.
func NewComponent() *ASAPTokenComponent {
	return &ASAPTokenComponent{}
}

// Settings generates a config populated with defaults.
func (m *ASAPTokenComponent) Settings() *ASAPTokenConfig {
	return &ASAPTokenConfig{
		TTL: time.Minute * -60,
	}
}

// New generates the middleware.
func (*ASAPTokenComponent) New(ctx context.Context, conf *ASAPTokenConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	if len(conf.PrivateKey) < 1 {
		return nil, fmt.Errorf("private key value is empty")
	}
	if len(conf.Issuer) < 1 {
		return nil, fmt.Errorf("issuer value is empty")
	}
	if len(conf.Audiences) < 1 {
		return nil, fmt.Errorf("audience list is empty")
	}
	if len(conf.KID) < 1 {
		return nil, fmt.Errorf("kid value is empty")
	}
	if conf.TTL == 0 || conf.TTL < 0 {
		return nil, fmt.Errorf("ttl duration is invalid: %s", conf.TTL)
	}
	rawKey := conf.PrivateKey
	if strings.HasPrefix(rawKey, "data:") {
		url, _ := dataurl.DecodeString(rawKey)
		rawKey = string(url.Data)
	}
	privateKey, err := asap.NewPrivateKey([]byte(rawKey))
	if err != nil {
		return nil, err
	}
	return asap.NewTransportDecorator(
		asap.NewCachingProvisioner(
			asap.NewProvisioner(conf.KID, conf.TTL, conf.Issuer, conf.Audiences, crypto.SigningMethodRS256),
		),
		privateKey,
	), nil
}
