package components

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTimeoutContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapped := NewMockRoundTripper(ctrl)
	rt := &timeoutRoundTripper{
		RoundTripper: wrapped,
		after:        time.Hour,
	}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	var capturedCtx context.Context

	wrapped.EXPECT().RoundTrip(gomock.Any()).DoAndReturn(
		func(r *http.Request) (*http.Response, error) {
			capturedCtx = r.Context()
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	)
	_, _ = rt.RoundTrip(req)
	assert.NotNil(t, capturedCtx)
	assert.Nil(t, capturedCtx.Err())
}
