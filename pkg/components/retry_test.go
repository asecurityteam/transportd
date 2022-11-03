package components

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRetryConfig(t *testing.T) {
	t.Parallel()

	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	rtComponent, err := Retry(context.Background(), "a", "b", "c")
	assert.Nil(t, err)
	config := rtComponent.(*RetryComponent).Settings()
	assert.Equal(t, "retry", config.Name())
	assert.Equal(t, 3, config.Limit)
	assert.Equal(t, false, config.Exponential)
	_, _ = rtComponent.(*RetryComponent).New(context.Background(), config)
	config.Exponential = true
	_, _ = rtComponent.(*RetryComponent).New(context.Background(), config)

}
