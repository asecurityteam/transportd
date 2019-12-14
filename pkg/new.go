package transportd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/asecurityteam/runhttp"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

type contextKey string

func (c contextKey) String() string {
	return "transportd_context_key_" + string(c)
}

var (
	// ContextKeyOpenAPISpec is a key used for placing the raw openapi3.Swagger object
	// into the context.  If it is needed in a component that defines a transportd plugin,
	// the object can be retrieved from the context passed to the "New" function by:
	// ctx.Value(transportd.ContextKeyOpenAPISpec).(*openapi3.Swagger)
	ContextKeyOpenAPISpec = contextKey("OpenAPISpec")
)

func newSpecification(source []byte) (*openapi3.Swagger, error) {
	envProcessor := NewEnvProcessor()
	source, err := envProcessor.Process(source)
	if err != nil {
		return nil, err
	}

	loader := openapi3.NewSwaggerLoader()
	swagger, errYaml := loader.LoadSwaggerFromData(source)
	var errJSON error
	if errYaml != nil {
		swagger, errJSON = loader.LoadSwaggerFromData(source)
	}
	if errYaml != nil && errJSON != nil {
		return nil, errJSON
	}
	return swagger, nil
}

func newTransport(ctx context.Context, specification *openapi3.Swagger, components ...NewComponent) (http.RoundTripper, error) {
	router := openapi3filter.NewRouter()
	err := router.AddSwagger(specification)
	if err != nil {
		return nil, err
	}

	// Load and configure available backends.
	var rawBackendConf interface{}
	var ok bool
	if rawBackendConf, ok = specification.Extensions[ExtensionKey]; !ok {
		return nil, fmt.Errorf("missing backend configuration")
	}
	s, err := SourceFromExtension([]byte(rawBackendConf.(json.RawMessage)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %s", err.Error())
	}
	transports, err := NewBaseTransports(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("failed to configure backends: %s", err.Error())
	}

	// Load and configure endpoints.
	reg := NewStaticClientRegistry()
	clientF := &ClientFactory{
		Bases:      transports,
		Components: components,
	}
	for path, pathItem := range specification.Paths {
		for method, op := range pathItem.Operations() {
			if _, ok = op.Extensions[ExtensionKey]; !ok {
				return nil, fmt.Errorf("missing client configuration for %s.%s", path, method)
			}
			opS, opErr := SourceFromExtension([]byte(op.Extensions[ExtensionKey].(json.RawMessage)))
			if opErr != nil {
				return nil, fmt.Errorf("failed to parse client configuration for %s.%s: %s", path, method, opErr.Error())
			}
			client, opErr := clientF.New(ctx, opS, path, method)
			if opErr != nil {
				return nil, fmt.Errorf("failed client configuration for %s.%s: %s", path, method, opErr.Error())
			}
			reg.Store(ctx, path, method, client)
		}
	}
	return &ClientTransport{
		Router:   router,
		Registry: reg,
	}, nil
}

// NewTransport constructs a smart HTTP client from the given specification
// and set of plugins. For running a service, use the New method instead.
func NewTransport(ctx context.Context, specification []byte, components ...NewComponent) (http.RoundTripper, error) {
	spec, err := newSpecification(specification)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, ContextKeyOpenAPISpec, spec)
	transport, err := newTransport(ctx, spec, components...)
	return transport, err
}

// New generates a configured HTTP runtime. To use as a library, call the
// NewTransport method instead.
func New(ctx context.Context, specification []byte, components ...NewComponent) (*runhttp.Runtime, error) {
	spec, err := newSpecification(specification)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, ContextKeyOpenAPISpec, spec)
	transport, err := newTransport(ctx, spec, components...)
	if err != nil {
		return nil, err
	}
	handler := &httputil.ReverseProxy{
		Director:  func(*http.Request) {},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			runhttp.LoggerFromContext(r.Context()).Error(struct {
				Message string `logevent:"message,default=uncaught-exception"`
				Reason  string `logevent:"reason"`
			}{
				Reason: err.Error(),
			})
			b, _ := json.Marshal(HTTPError{
				Code:   http.StatusBadGateway,
				Status: http.StatusText(http.StatusBadGateway),
				Reason: err.Error(),
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write(b)
		},
		ModifyResponse: (MultiResponseModifier{
			EnforceRelativeLocation,
		}).ModifyResponse,
	}

	// Load and configure the runtime settings.
	var rawRuntimeConf interface{}
	var ok bool
	if rawRuntimeConf, ok = spec.Extensions[RuntimeExtensionKey]; !ok {
		return nil, fmt.Errorf("missing x-runtime configuration")
	}
	s, err := RuntimeSourceFromExtension([]byte(rawRuntimeConf.(json.RawMessage)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime configuration: %s", err.Error())
	}
	rt, err := NewRuntime(ctx, s, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to configure runtime: %s", err.Error())
	}
	return rt, nil
}
