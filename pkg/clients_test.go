package transportd

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	cfComponentName = "cfComponent"
)

type cfComponentConfig struct {
	V int
}

func (*cfComponentConfig) Name() string {
	return cfComponentName
}

type cfComponent struct {
	Conf     *cfComponentConfig
	Err      error
	AdaptErr error
}

func (*cfComponent) Settings() *cfComponentConfig {
	return &cfComponentConfig{V: 2}
}
func (c *cfComponent) New(_ context.Context, conf *cfComponentConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	c.Conf = conf
	return func(w http.RoundTripper) http.RoundTripper {
		return w
	}, c.Err
}
func (c *cfComponent) Adapt(_ context.Context, backend string, path string, method string) (interface{}, error) {
	return c, c.AdaptErr
}

func TestNewClientFactoryFailedToLoadEnabledList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	comp := &cfComponent{}
	br := NewMockBackendRegistry(ctrl)
	s := NewMockSource(ctrl)
	cf := &ClientFactory{
		Bases:      br,
		Components: []NewComponent{comp.Adapt},
	}

	s.EXPECT().Get(ctx, ExtensionKey, enabledSetting).Return(nil, true)
	_, err := cf.New(ctx, s, "", "")
	assert.NotNil(t, err)
}

func TestNewClientFactoryFailedMissingBackend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	comp := &cfComponent{}
	br := NewMockBackendRegistry(ctrl)
	s := NewMockSource(ctrl)
	cf := &ClientFactory{
		Bases:      br,
		Components: []NewComponent{comp.Adapt},
	}

	s.EXPECT().Get(ctx, ExtensionKey, enabledSetting).Return([]string{cfComponentName}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backendSetting).Return("b", true)
	br.EXPECT().Load(ctx, "b").Return(nil)
	_, err := cf.New(ctx, s, "", "")
	assert.NotNil(t, err)
}

func TestNewClientFactoryFailedComponentFactoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	rt := NewMockBackend(ctrl)
	comp := &cfComponent{AdaptErr: errors.New("")}
	br := NewMockBackendRegistry(ctrl)
	s := NewMockSource(ctrl)
	cf := &ClientFactory{
		Bases:      br,
		Components: []NewComponent{comp.Adapt},
	}

	s.EXPECT().Get(ctx, ExtensionKey, enabledSetting).Return([]string{cfComponentName}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backendSetting).Return("", true)
	br.EXPECT().Load(ctx, "").Return(rt)
	_, err := cf.New(ctx, s, "", "")
	assert.NotNil(t, err)
}

func TestNewClientFactoryFailedComponentLoadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	rt := NewMockBackend(ctrl)
	comp := &cfComponent{Err: errors.New("")}
	br := NewMockBackendRegistry(ctrl)
	s := NewMockSource(ctrl)
	cf := &ClientFactory{
		Bases:      br,
		Components: []NewComponent{comp.Adapt},
	}

	s.EXPECT().Get(ctx, ExtensionKey, enabledSetting).Return([]string{cfComponentName}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backendSetting).Return("", true)
	br.EXPECT().Load(ctx, "").Return(rt)
	s.EXPECT().Get(ctx, ExtensionKey, cfComponentName, "V").Return("a", true)
	_, err := cf.New(ctx, s, "", "")
	assert.NotNil(t, err)
}

func TestNewClientFactoryFailedComponentMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	rt := NewMockBackend(ctrl)
	u, _ := url.Parse("https://127.0.0.1")
	comp := &cfComponent{}
	br := NewMockBackendRegistry(ctrl)
	s := NewMockSource(ctrl)
	cf := &ClientFactory{
		Bases:      br,
		Components: []NewComponent{comp.Adapt},
	}

	s.EXPECT().Get(ctx, ExtensionKey, enabledSetting).Return([]string{cfComponentName, "t"}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backendSetting).Return("", true)
	br.EXPECT().Load(ctx, "").Return(rt)
	rt.EXPECT().Host().Return(u).AnyTimes()
	_, err := cf.New(ctx, s, "", "")
	assert.NotNil(t, err)
}

func TestNewClientFactorySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	rt := NewMockBackend(ctrl)
	u, _ := url.Parse("https://127.0.0.1")
	comp := &cfComponent{}
	br := NewMockBackendRegistry(ctrl)
	s := NewMockSource(ctrl)
	cf := &ClientFactory{
		Bases:      br,
		Components: []NewComponent{comp.Adapt},
	}

	s.EXPECT().Get(ctx, ExtensionKey, enabledSetting).Return([]string{cfComponentName}, true)
	s.EXPECT().Get(ctx, ExtensionKey, backendSetting).Return("", true)
	br.EXPECT().Load(ctx, "").Return(rt)
	rt.EXPECT().Host().Return(u).AnyTimes()
	s.EXPECT().Get(ctx, ExtensionKey, cfComponentName, "V").Return(1, true)
	client, err := cf.New(ctx, s, "", "")
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, comp.Conf.V, 1)
}
