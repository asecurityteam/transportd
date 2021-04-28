package components

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asecurityteam/logevent"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func simpleResponse() *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       http.NoBody,
	}
}

func TestAccessLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockLogger(ctrl)
	rt := NewMockRoundTripper(ctrl)

	req := httptest.NewRequest(http.MethodGet, "https://localhost/", http.NoBody)
	req = req.WithContext(
		context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}),
	)
	req = req.WithContext(logevent.NewContext(req.Context(), logger))
	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		assert.IsType(t, accessLog{}, event, "middleware did not perform an access log")
	})
	rt.EXPECT().RoundTrip(gomock.Any()).Return(simpleResponse(), nil).AnyTimes()
	wrapped := &loggingTransport{
		Wrapped: rt,
	}
	_, _ = wrapped.RoundTrip(req)
}

func TestPrincipalLogging(t *testing.T) {
	tests := []struct {
		name              string
		expectedHeader    string
		sentHeader        string
		expectedPrincipal string
		sentPrincipal     string
	}{
		{
			name:              "Found Header",
			expectedHeader:    "SUBJECT-HEADER",
			sentHeader:        "SUBJECT-HEADER",
			expectedPrincipal: "some-user",
			sentPrincipal:     "some-user",
		},
		{
			name:              "Expected wrong Header",
			expectedHeader:    "X-Slauth-Subject",
			sentHeader:        "SUBJECT-HEADER",
			expectedPrincipal: "",
			sentPrincipal:     "some-user",
		},
	}

	//Base request and setup that doesn't need to change
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockLogger(ctrl)
	rt := NewMockRoundTripper(ctrl)

	req := httptest.NewRequest(http.MethodGet, "https://localhost/", http.NoBody)
	req = req.WithContext(
		context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}),
	)
	req = req.WithContext(logevent.NewContext(req.Context(), logger))
	rt.EXPECT().RoundTrip(gomock.Any()).Return(simpleResponse(), nil).AnyTimes()

	//Tests for expectations
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req.Header.Add(tt.sentHeader, tt.sentPrincipal)

			logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
				assert.IsType(t, accessLog{}, event, "middleware did not perform an access log")
				//Safe because we've already verified type
				log := event.(accessLog)
				assert.Equal(t, tt.expectedPrincipal, log.Principal)
			})

			wrapped := &loggingTransport{
				Wrapped:         rt,
				PrincipalHeader: tt.expectedHeader,
			}
			_, _ = wrapped.RoundTrip(req)

			//Since we're reusing requests instead of remaking them each run
			req.Header.Del(tt.sentHeader)
		})
	}
}
