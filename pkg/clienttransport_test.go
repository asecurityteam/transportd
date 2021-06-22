package transportd

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	spectxt = `
    openapi: 3.0.0
    info:
      version: 1.0.0
      title: Hello API
      description: A hello world API.
      termsOfService: 'http://swagger.io/terms/'
      contact:
        name: Swagger API Team
        email: apiteam@swagger.io
        url: 'http://swagger.io'
      license:
        name: Apache 2.0
        url: 'https://www.apache.org/licenses/LICENSE-2.0.html'
    paths:
      /hello:
        get:
          description: Fetches a greeting.
          operationId: hello
          parameters:
            - name: name
              in: query
              description: name of person being greeted
              required: true
              schema:
                type: string
          responses:
            '200':
              description: hello response
              content:
                application/json:
                  schema:
                    $ref: '#/components/schemas/Greeting'
            default:
              description: unexpected error
              content:
                application/text:
                  schema:
                    type: string
    components:
      schemas:
        Greeting:
          required:
            - greeting
          properties:
            greeting:
              type: string
    `
)

func TestClientTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := NewMockRoundTripper(ctrl)
	reg := NewMockClientRegistry(ctrl)
	loader := openapi3.NewLoader()
	spec, _ := loader.LoadFromData([]byte(spectxt))
	router, err := legacyrouter.NewRouter(spec)
	assert.Nil(t, err)
	rt := &ClientTransport{
		Registry: reg,
		Router:   router,
	}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/hello", http.NoBody)

	reg.EXPECT().Load(gomock.Any(), "/hello", http.MethodGet).Return(cl)
	cl.EXPECT().RoundTrip(gomock.Any()).Do(func(r *http.Request) {
		route := RouteFromContext(r.Context())
		assert.Equal(t, "/hello", route.Path)
		assert.Equal(t, http.MethodGet, route.Method)
	})
	_, err = rt.RoundTrip(req)
	assert.Nil(t, err)
}

func TestClientTransportNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reg := NewMockClientRegistry(ctrl)
	loader := openapi3.NewLoader()
	spec, _ := loader.LoadFromData([]byte(spectxt))
	router, err := legacyrouter.NewRouter(spec)
	assert.Nil(t, err)
	rt := &ClientTransport{
		Registry: reg,
		Router:   router,
	}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/something", http.NoBody)

	reg.EXPECT().Load(gomock.Any(), unknownKey, unknownKey).Return(nil)
	resp, err := rt.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestClientTransportPassthrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := NewMockRoundTripper(ctrl)
	reg := NewMockClientRegistry(ctrl)
	loader := openapi3.NewLoader()
	spec, _ := loader.LoadFromData([]byte(spectxt))
	router, err := legacyrouter.NewRouter(spec)
	assert.Nil(t, err)
	rt := &ClientTransport{
		Registry: reg,
		Router:   router,
	}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/something", http.NoBody)
	reg.EXPECT().Load(gomock.Any(), unknownKey, unknownKey).Return(cl)
	cl.EXPECT().RoundTrip(gomock.Any()).Do(func(r *http.Request) {
		route := RouteFromContext(r.Context())
		assert.Equal(t, unknownKey, route.Path)
		assert.Equal(t, http.MethodGet, route.Method)
	})

	_, err = rt.RoundTrip(req)
	assert.Nil(t, err)
}
