// +build integration

package inttest

import (
	"context"
	"testing"

	"github.com/asecurityteam/transportd/pkg"
	"github.com/asecurityteam/transportd/pkg/components"
	"github.com/stretchr/testify/assert"
)

func TestHelpNoErrors(t *testing.T) {
	h, err := transportd.Help(
		context.Background(),
		components.Metrics,
		components.AccessLog,
		components.ASAPValidate,
		components.Timeout,
		components.Hedging,
		components.Retry,
		components.ASAPToken,
		components.Strip,
		components.RequestValidation,
		components.ResponseValidation,
	)
	// Basic sanity check that the help output works with
	// the native components and is not empty.
	assert.Nil(t, err)
	assert.NotEmpty(t, h)
}
