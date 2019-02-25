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
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       http.NoBody,
	}
	rt.EXPECT().RoundTrip(gomock.Any()).Return(resp, nil).AnyTimes()
	wrapped := &loggingTransport{
		Wrapped: rt,
	}
	_, _ = wrapped.RoundTrip(req)
}
