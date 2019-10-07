package transportd

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	testBackend = "test"
)

func TestHostRewrite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapped := NewMockRoundTripper(ctrl)
	rt := &hostRewrite{
		Scheme:  "https",
		Host:    "test",
		Wrapped: wrapped,
	}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)

	wrapped.EXPECT().RoundTrip(gomock.Any()).Do(func(req *http.Request) {
		assert.Equal(t, req.Host, rt.Host)
		assert.Equal(t, req.URL.Host, rt.Host)
		assert.Equal(t, req.URL.Scheme, rt.Scheme)
		assert.Equal(t, req.RequestURI, "")
		assert.Equal(t, req.URL.Opaque, "")
		assert.Equal(t, req.URL.RawPath, "")
	}).Return(nil, nil)
	_, _ = rt.RoundTrip(req)
}

func Test_validateHost(t *testing.T) {
	tests := []struct {
		name    string
		u       string
		wantErr bool
	}{
		{
			name:    "missing scheme",
			u:       `localhost`,
			wantErr: true,
		},
		{
			name:    "missing host",
			u:       `https://`,
			wantErr: true,
		},
		{
			name:    "passing",
			u:       `https://localhost`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse(tt.u)
			if err := validateHost(u); (err != nil) != tt.wantErr {
				t.Errorf("validateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBaseTransportFailToLoadBackendsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	s := NewMockSource(ctrl)

	s.EXPECT().Get(ctx, ExtensionKey, backendsSetting).Return(nil, true)
	_, err := NewBaseTransports(ctx, s)
	assert.NotNil(t, err)
}

func TestNewBaseTransportFailToLoadBackendItemPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	s := NewMockSource(ctrl)
	backend1 := testBackend

	s.EXPECT().Get(ctx, ExtensionKey, backendsSetting).Return([]string{backend1}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, hostSetting).Return("http://localhost", true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, poolSetting, countSetting).Return("a", true)
	_, err := NewBaseTransports(ctx, s)
	assert.NotNil(t, err)
}

func TestNewBaseTransportFailToLoadBackendItemHost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	s := NewMockSource(ctrl)
	backend1 := testBackend

	s.EXPECT().Get(ctx, ExtensionKey, backendsSetting).Return([]string{backend1}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, hostSetting).Return("", true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, poolSetting, countSetting).Return(1, true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, poolSetting, ttlSetting).Return(time.Hour, true)
	_, err := NewBaseTransports(ctx, s)
	assert.NotNil(t, err)
}

func TestNewBaseTransportSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	s := NewMockSource(ctrl)
	backend1 := testBackend

	s.EXPECT().Get(ctx, ExtensionKey, backendsSetting).Return([]string{backend1}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, hostSetting).Return("https://localhost", true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, poolSetting, countSetting).Return(1, true)
	s.EXPECT().Get(ctx, ExtensionKey, backend1, poolSetting, ttlSetting).Return(time.Hour, true)
	result, err := NewBaseTransports(ctx, s)
	assert.Nil(t, err)
	assert.NotNil(t, result.Load(ctx, backend1))
}
