package components

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"testing"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	validatorYaml = `
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
    parameters:
      - name: name2
        in: query
        description: a test variable in the path item
        required: true
        schema:
          type: string
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
            text/plain:
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

func TestValidateRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router := openapi3filter.NewRouter()
	assert.Nil(t, router.AddSwagger(swagger))

	rt := NewMockRoundTripper(ctrl)
	c := &inputValidatingTransport{Wrapped: rt}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1&name2=test2", http.NoBody)

	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	assert.Nil(t, err)
	req = req.WithContext(transportd.RouteToContext(req.Context(), route))
	req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

	rt.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}, nil)
	_, err = c.RoundTrip(req)
	assert.Nil(t, err)
}

func TestValidateRequestMissingParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router := openapi3filter.NewRouter()
	assert.Nil(t, router.AddSwagger(swagger))

	rt := NewMockRoundTripper(ctrl)
	c := &inputValidatingTransport{Wrapped: rt}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1", http.NoBody)

	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	assert.Nil(t, err)
	req = req.WithContext(transportd.RouteToContext(req.Context(), route))
	req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

	resp, err := c.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestValidateResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router := openapi3filter.NewRouter()
	assert.Nil(t, router.AddSwagger(swagger))

	rt := NewMockRoundTripper(ctrl)
	c := &outputValidatingTransport{Wrapped: rt}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1&name2=test2", http.NoBody)

	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	assert.Nil(t, err)
	req = req.WithContext(transportd.RouteToContext(req.Context(), route))
	req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

	rt.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"greeting": "hello"}`)),
	}, nil)
	resp, err := c.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateResponseMissingHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router := openapi3filter.NewRouter()
	assert.Nil(t, router.AddSwagger(swagger))

	rt := NewMockRoundTripper(ctrl)
	c := &outputValidatingTransport{Wrapped: rt}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1&name2=test2", http.NoBody)

	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	assert.Nil(t, err)
	req = req.WithContext(transportd.RouteToContext(req.Context(), route))
	req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

	rt.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"greeting": "hello"}`)),
	}, nil)
	resp, err := c.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestValidatorResponseBadShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router := openapi3filter.NewRouter()
	assert.Nil(t, router.AddSwagger(swagger))

	rt := NewMockRoundTripper(ctrl)
	c := &outputValidatingTransport{Wrapped: rt}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1&name2=test2", http.NoBody)

	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	assert.Nil(t, err)
	req = req.WithContext(transportd.RouteToContext(req.Context(), route))
	req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

	rt.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"notagreeting": "hello"}`)),
	}, nil)
	resp, err := c.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestValidateCompressedResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router := openapi3filter.NewRouter()
	assert.Nil(t, router.AddSwagger(swagger))

	rt := NewMockRoundTripper(ctrl)
	c := &outputValidatingTransport{Wrapped: rt}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1&name2=test2", http.NoBody)
	req.Header.Set("Accept-Encoding", "gzip")
	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	assert.Nil(t, err)
	req = req.WithContext(transportd.RouteToContext(req.Context(), route))
	req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write([]byte(`{"greeting": "hello"}`))
	assert.Nil(t, err)
	assert.Nil(t, zw.Close())

	rt.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header: http.Header{
			"Content-Type":     []string{"application/json"},
			"Content-Encoding": []string{"gzip"},
		},
		Body: ioutil.NopCloser(&buf),
	}, nil)
	resp, err := c.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
