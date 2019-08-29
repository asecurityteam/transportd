package components

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func Test_incomingMatchesAllowed(t *testing.T) {
	defaultAllowedHeaderAndValues := map[string][]string{"Ldap-Groups": {"sre", "devs"}}
	tests := []struct {
		name           string
		allowedHeader  map[string][]string
		incomingHeader map[string][]string
		wantResult     bool
		wantErr        bool
	}{
		{
			name:           "multiple incoming header values with multiple allowed values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"sre", "devs"}},
			wantResult:     true,
			wantErr:        false,
		},
		{
			name:           "multiple incoming header values with a single allowed value",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"devs"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"sre", "devs"}},
			wantResult:     true,
			wantErr:        false,
		},
		{
			name:           "single incoming header value with multiple allowed values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"devs"}},
			wantResult:     true,
			wantErr:        false,
		},
		{
			name:           "single incoming header value with a single allowed value",
			allowedHeader:  map[string][]string{"Ldap-Groups": {"devs"}},
			incomingHeader: map[string][]string{"Ldap-Groups": {"devs"}},
			wantResult:     true,
			wantErr:        false,
		},
		{
			name:           "incorrect incoming header value",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"Ldap-Groups": {"design"}},
			wantResult:     false,
			wantErr:        true,
		},
		{
			name:           "missing specific incoming header",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"meh": {"sre", "devs"}},
			wantResult:     false,
			wantErr:        true,
		},
		{
			name:           "missing incoming header and values",
			allowedHeader:  defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"": {""}},
			wantResult:     false,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := incomingMatchesRequired(tt.allowedHeader, tt.incomingHeader)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_contains(t *testing.T) {
	defaultS := []string{"dog", "cat", "bird", "fish"}
	tests := []struct {
		name         string
		sliceToCheck []string
		target       string
		wantResult   bool
	}{
		{
			name:         "target is not present",
			sliceToCheck: defaultS,
			target:       "insect",
			wantResult:   false,
		},
		{
			name:         "target is empty",
			sliceToCheck: defaultS,
			target:       "",
			wantResult:   false,
		},
		{
			name:         "target is present",
			sliceToCheck: defaultS,
			target:       "fish",
			wantResult:   true,
		},
		{
			name:         "target is present in a single value slice",
			sliceToCheck: []string{"dog"},
			target:       "dog",
			wantResult:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, contains(tt.sliceToCheck, tt.target), tt.wantResult)
		})
	}
}

func Test_validateHeadersRoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		allowedHeaders map[string][]string
		testHeaders    http.Header
		wantErr        bool
		wantResponse   int
	}{
		{
			name:           "valid header and values present",
			allowedHeaders: map[string][]string{"client": {"mobile"}},
			testHeaders: http.Header{
				"Client": {"mobile", "browser"},
			},
			wantErr:      false,
			wantResponse: http.StatusOK,
		},
		{
			name:           "missing allowed header value",
			allowedHeaders: map[string][]string{"client": {"browser"}},
			testHeaders: http.Header{
				"Client": {"telnet"},
			},
			wantErr:      false,
			wantResponse: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rt := NewMockRoundTripper(ctrl)
			c := &validateHeaderTransport{
				Wrapped:  rt,
				Required: tt.allowedHeaders,
			}
			r := &http.Request{Header: tt.testHeaders}
			rt.EXPECT().RoundTrip(gomock.Any()).Return(
				&http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				},
				nil,
			).AnyTimes()
			got, err := c.RoundTrip(r)
			require.Equal(t, err != nil, tt.wantErr)
			require.Equal(t, tt.wantResponse, got.StatusCode)
		})
	}
}
