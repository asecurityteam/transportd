package transportd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorToStatusCode(t *testing.T) {
	code := ErrorToStatusCode(context.Canceled)
	assert.Equal(t, 504, code)

	code = ErrorToStatusCode(context.DeadlineExceeded)
	assert.Equal(t, 504, code)

	code = ErrorToStatusCode(nil)
	assert.Equal(t, 502, code)
}
