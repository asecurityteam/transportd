package components

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestRequestHeaderInjection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rt := NewMockRoundTripper(ctrl)
	cmp := &RequestHeaderComponent{}
	set := cmp.Settings()
	set.Headers = map[string][]string{"one": {"a"}, "two": {"b"}, "three": {"c"}}
	d, err := cmp.New(context.Background(), set)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	wrapped := d(rt)

	rt.EXPECT().RoundTrip(gomock.Any()).Do(func(r *http.Request) {
		assert.Equal(t, "a", r.Header.Get("one"))
		assert.Equal(t, "b", r.Header.Get("two"))
		assert.Equal(t, "c", r.Header.Get("three"))
	})
	_, err = wrapped.RoundTrip(req)
	assert.NoError(t, err)
}

func TestResponseHeaderInjection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rt := NewMockRoundTripper(ctrl)
	cmp := &ResponseHeaderComponent{}
	set := cmp.Settings()
	set.Headers = map[string][]string{"one": {"a"}, "two": {"b"}, "three": {"c"}}
	d, err := cmp.New(context.Background(), set)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	wrapped := d(rt)

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header: http.Header{
			"One":  []string{"NOT A"},
			"Four": []string{"d"},
		},
		Body: io.NopCloser(bytes.NewReader([]byte(""))),
	}

	rt.EXPECT().RoundTrip(gomock.Any()).Return(mockResponse, nil)
	resp, err := wrapped.RoundTrip(req)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, "a", resp.Header.Get("one"))
	assert.Equal(t, "b", resp.Header.Get("two"))
	assert.Equal(t, "c", resp.Header.Get("three"))
	assert.Equal(t, "d", resp.Header.Get("four"))

}
