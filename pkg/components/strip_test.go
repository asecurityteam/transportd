package components

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStripTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rt := NewMockRoundTripper(ctrl)
	c := &strippingTransport{
		Wrapped: rt,
		Count:   2,
	}
	req, _ := http.NewRequest(http.MethodGet, "https://localhost/one/two/three/four", http.NoBody)

	rt.EXPECT().RoundTrip(gomock.Any()).Do(func(r *http.Request) {
		assert.Equal(t, "https://localhost/three/four", r.URL.String())
	}).Return(nil, nil)
	_, _ = c.RoundTrip(req)
}
