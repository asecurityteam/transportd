package transportd

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestClientRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reg := NewStaticClientRegistry()
	ctx := context.Background()
	rt := NewMockRoundTripper(ctrl)

	// Nil when not set
	assert.Nil(t, reg.Load(ctx, "a", "b"))

	// Not nil and case insensitive when set
	reg.Store(ctx, "A", "B", rt)
	assert.Equal(t, rt, reg.Load(ctx, "a", "b"))
}
