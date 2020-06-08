package components

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRetryAfterConfig(t *testing.T) {
	t.Parallel()

	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	rtComponent, err := RetryAfter(context.Background(), "a", "b", "c")
	assert.Nil(t, err)
	config := rtComponent.(*RetryAfterComponent).Settings()
	assert.Equal(t, "retryafter", config.Name())
	_, _ = rtComponent.(*RetryAfterComponent).New(context.Background(), config)

}
