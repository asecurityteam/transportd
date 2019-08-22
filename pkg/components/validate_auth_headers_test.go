package components

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_incomingMatchesAllowed(t *testing.T) {
	allowedHeader := make(map[string][]string)
	goodIncomingHeader := make(map[string][]string)
	badIncomingHeader := make(map[string][]string)

	goodIncomingHeader["LDAP-Groups"] = []string{"sre", "devs"}
	badIncomingHeader["LDAP-Groups"] = []string{"design"}
	allowedHeader["LDAP-Groups"] = []string{"sre"}

	result := incomingMatchesAllowed(allowedHeader, goodIncomingHeader)
	assert.Equal(t, result, true)
	result = incomingMatchesAllowed(allowedHeader, badIncomingHeader)
	assert.Equal(t, result, false)
	badIncomingHeader["LDAP-Groups"] = nil
	result = incomingMatchesAllowed(allowedHeader, badIncomingHeader)
	assert.Equal(t, result, false)
}