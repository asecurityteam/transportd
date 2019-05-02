package components

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestHeaderInjectionConfigError(t *testing.T) {
	cmp := &HeaderComponent{}
	set := cmp.Settings()
	set.Names = []string{"one", "two"}
	set.Values = []string{"one"}
	_, err := cmp.New(context.Background(), set)
	assert.Error(t, err)
}

func TestHeaderInjection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rt := NewMockRoundTripper(ctrl)
	cmp := &HeaderComponent{}
	set := cmp.Settings()
	set.Names = []string{"one", "two", "three"}
	set.Values = []string{"a", "b", "c"}
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
