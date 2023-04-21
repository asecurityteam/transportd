package components

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"testing"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
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
      security:
        - BearerAuth: []
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
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
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
	tests := []struct {
		name        string
		url         string
		response    *http.Response
		responseErr error
		statusCode  int
	}{
		{
			name: "valid request",
			url:  "https://localhost/hello?name=test1&name2=test2",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			},
			responseErr: nil,
			statusCode:  http.StatusOK,
		},
		{
			name:        "missing param",
			url:         "https://localhost/hello?name=test1",
			response:    nil,
			responseErr: nil,
			statusCode:  http.StatusBadRequest,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router, err := legacyrouter.NewRouter(swagger)
	assert.Nil(t, err)

	for _, tt := range tests {
		rt := NewMockRoundTripper(ctrl)
		c := &inputValidatingTransport{Wrapped: rt}
		req, _ := http.NewRequest(http.MethodGet, tt.url, http.NoBody)

		route, pathParams, err := router.FindRoute(req)
		assert.Nil(t, err)
		req = req.WithContext(transportd.RouteToContext(req.Context(), route))
		req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

		if tt.response != nil {
			rt.EXPECT().RoundTrip(gomock.Any()).Return(tt.response, tt.responseErr)
		}
		resp, err := c.RoundTrip(req)
		assert.Nil(t, err)
		assert.Equal(t, tt.statusCode, resp.StatusCode)
	}
}

func TestValidateResponse(t *testing.T) {
	body := `{"greeting": "hello"}`
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(body))
	assert.Nil(t, err)
	assert.Nil(t, zw.Close())
	compressedBody, err := io.ReadAll(&buf)
	assert.Nil(t, err)

	tests := []struct {
		name        string
		response    *http.Response
		responseErr error
		expectedErr bool
		statusCode  int
	}{
		{
			name: "valid response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			},
			responseErr: nil,
			expectedErr: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "missing header",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			},
			responseErr: nil,
			expectedErr: false,
			statusCode:  http.StatusBadGateway,
		},
		{
			name: "compressed response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header: http.Header{
					"Content-Type":     []string{"application/json"},
					"Content-Encoding": []string{"gzip"},
				},
				Body: io.NopCloser(bytes.NewReader(compressedBody)),
			},
			responseErr: nil,
			expectedErr: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "compressed response missing header",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(bytes.NewReader(compressedBody)),
			},
			responseErr: nil,
			expectedErr: false,
			statusCode:  http.StatusBadGateway,
		},
		{
			name: "compressed header with uncompressed body",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header: http.Header{
					"Content-Type":     []string{"application/json"},
					"Content-Encoding": []string{"gzip"},
				},
				Body: io.NopCloser(bytes.NewBufferString(body)),
			},
			responseErr: nil,
			expectedErr: true,
			statusCode:  -1,
		},
		{
			name:        "response error",
			response:    nil,
			responseErr: fmt.Errorf("response error"),
			expectedErr: true,
			statusCode:  -1,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(validatorYaml))
	assert.Nil(t, err)
	router, err := legacyrouter.NewRouter(swagger)
	assert.Nil(t, err)

	for _, tt := range tests {
		rt := NewMockRoundTripper(ctrl)
		c := &outputValidatingTransport{Wrapped: rt}
		req, _ := http.NewRequest(http.MethodGet, "https://localhost/hello?name=test1&name2=test2", http.NoBody)
		req.Header.Set("Accept-Encoding", "gzip")
		route, pathParams, err := router.FindRoute(req)
		assert.Nil(t, err)
		req = req.WithContext(transportd.RouteToContext(req.Context(), route))
		req = req.WithContext(transportd.PathParamsToContext(req.Context(), pathParams))

		rt.EXPECT().RoundTrip(gomock.Any()).Return(tt.response, tt.responseErr)
		resp, err := c.RoundTrip(req)
		if tt.expectedErr {
			assert.NotNil(t, err)
			assert.Nil(t, resp)
			return
		}
		assert.Nil(t, err)
		assert.Equal(t, tt.statusCode, resp.StatusCode)
	}
}
