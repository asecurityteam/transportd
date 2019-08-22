package components

import (
	"testing"
)

func Test_incomingMatchesAllowed(t *testing.T) {
	defaultAllowedHeaderAndValues := map[string][]string{"LDAP-Groups": {"sre", "devs"}}
	tests := []struct {
		name string
		allowedHeader map[string][]string
		incomingHeader map[string][]string
		wantResult bool
	}{
		{
			name: "good incoming header values",
			allowedHeader: defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"LDAP-Groups": {"sre", "devs"}},
			wantResult: true,
		},
		{
			name: "good incoming header values with single required header value",
			allowedHeader: map[string][]string{"LDAP-Groups": {"devs"}},
			incomingHeader: map[string][]string{"LDAP-Groups": {"sre", "devs"}},
			wantResult: true,
		},
		{
			name: "bad incoming header values",
			allowedHeader: defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"LDAP-Groups": {"design"}},
			wantResult: false,
		},
		{
			name: "missing incoming header",
			allowedHeader: defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"meh": {"sre", "devs"}},
			wantResult: false,
		},
		{
			name: "missing single incoming header value",
			allowedHeader: defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"LDAP-Groups": {"devs"}},
			wantResult: false,
		},
		{
			name: "missing incoming header and values",
			allowedHeader: defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"": {""}},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			if result := incomingMatchesAllowed(tt.allowedHeader, tt.incomingHeader); result != tt.wantResult {
				t.Errorf("wanted: %t got: %t", tt.wantResult, result)
			}
		})
	}
}