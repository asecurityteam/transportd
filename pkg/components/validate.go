package components

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/getkin/kin-openapi/openapi3filter"
)

type inputValidatingTransport struct {
	Wrapped http.RoundTripper
}

func (r *inputValidatingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	route := transportd.RouteFromContext(req.Context())
	params := transportd.PathParamsFromContext(req.Context())
	input := &openapi3filter.RequestValidationInput{
		Route:       route,
		Request:     req,
		QueryParams: req.URL.Query(),
		PathParams:  params,
		Options: &openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error { return nil },
		},
	}
	err := openapi3filter.ValidateRequest(req.Context(), input)
	if err != nil {
		return newError(http.StatusBadRequest, err.Error()), nil
	}
	return r.Wrapped.RoundTrip(req)
}

// RequestValidationConfig is a placeholder for future validation options.
type RequestValidationConfig struct{}

// Name of the config root.
func (*RequestValidationConfig) Name() string {
	return "requestvalidation"
}

// RequestValidationComponent enables validation of requests against the
// OpenAPI specification.
type RequestValidationComponent struct{}

// RequestValidation satisfies the NewComponent signature.
func RequestValidation(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &RequestValidationComponent{}, nil
}

// Settings generates a config with all defaults set.
func (*RequestValidationComponent) Settings() *RequestValidationConfig {
	return &RequestValidationConfig{}
}

// New generates the middleware.
func (*RequestValidationComponent) New(_ context.Context, _ *RequestValidationConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &inputValidatingTransport{Wrapped: wrapped}
	}, nil
}

type outputValidatingTransport struct {
	Wrapped http.RoundTripper
}

func (r *outputValidatingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	originalPath := req.URL.Path
	resp, err := r.Wrapped.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// restore the old path just in case something else modified it from the path in the specification
	req.URL.Path = originalPath
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	route := transportd.RouteFromContext(req.Context())
	params := transportd.PathParamsFromContext(req.Context())
	reqInput := &openapi3filter.RequestValidationInput{
		Route:       route,
		Request:     req,
		QueryParams: req.URL.Query(),
		PathParams:  params,
	}

	// response validation will fail if the response body is compressed,
	// so we ensure the payload is uncompressed
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	default:
		reader = io.NopCloser(bytes.NewReader(body))
	}
	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqInput,
		Status:                 resp.StatusCode,
		Header:                 resp.Header,
		Body:                   reader,
	}
	err = openapi3filter.ValidateResponse(req.Context(), input)
	if err != nil {
		return newError(http.StatusBadGateway, err.Error()), nil
	}
	return resp, nil
}

// ResponseValidationConfig is a placeholder for future validation options.
type ResponseValidationConfig struct{}

// ResponseValidation satisfies the NewComponent signature.
func ResponseValidation(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &ResponseValidationComponent{}, nil
}

// Name of the config root.
func (*ResponseValidationConfig) Name() string {
	return "responsevalidation"
}

// ResponseValidationComponent enables validation of requests against the
// OpenAPI specification.
type ResponseValidationComponent struct{}

// Settings generates a config with all defaults set.
func (*ResponseValidationComponent) Settings() *ResponseValidationConfig {
	return &ResponseValidationConfig{}
}

// New generates the middleware.
func (*ResponseValidationComponent) New(_ context.Context, _ *ResponseValidationConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	return func(wrapped http.RoundTripper) http.RoundTripper {
		return &outputValidatingTransport{Wrapped: wrapped}
	}, nil
}
