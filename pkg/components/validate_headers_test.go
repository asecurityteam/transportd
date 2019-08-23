package components

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
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
			name: "good single incoming header value",
			allowedHeader: map[string][]string{"LDAP-Groups": {"devs"}},
			incomingHeader: map[string][]string{"LDAP-Groups": {"sre", "devs"}},
			wantResult: true,
		},
		{
			name: "incorrect incoming header value",
			allowedHeader: defaultAllowedHeaderAndValues,
			incomingHeader: map[string][]string{"LDAP-Groups": {"design"}},
			wantResult: false,
		},
		{
			name: "missing specific incoming header",
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
			assert.Equal(t, incomingMatchesAllowed(tt.allowedHeader, tt.incomingHeader), tt.wantResult)
		})
	}
}

func Test_contains(t *testing.T) {
	defaultS := []string{"dog", "cat", "bird", "fish"}
	tests := []struct {
		name string
		sliceToCheck []string
		target string
		wantResult bool
	}{
		{
			name: "target is not present",
			sliceToCheck: defaultS,
			target: "insect",
			wantResult: false,
		},
		{
			name: "target is empty",
			sliceToCheck: defaultS,
			target: "",
			wantResult: false,
		},
		{
			name: "target is present",
			sliceToCheck: defaultS,
			target: "fish",
			wantResult: true,
		},
		{
			name: "target is present in a single value slice",
			sliceToCheck: []string{"dog"},
			target: "dog",
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			assert.Equal(t, contains(tt.sliceToCheck, tt.target), tt.wantResult)
		})
	}
}

func Test_validateHeadersRoundTrip(t *testing.T) {
	tests := []struct{
		name string
		allowedHeaders map[string][]string
		testHeaders http.Header
		wantErr bool
		wantResponse int
	}{
		{
			name: "valid header and values present",
			allowedHeaders: map[string][]string{"client": {"browser", "mobile"}},
			testHeaders: http.Header{
				"client": {"mobile"},
			},
			wantErr: false,
			wantResponse: http.StatusOK,
		},
		{
			name: "missing allowed header value",
			allowedHeaders: map[string][]string{"client": {"browser", "mobile"}},
			testHeaders: http.Header{
				"client": {"telnet"},
			},
			wantErr: true,
			wantResponse: http.StatusUnauthorized,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rt := NewMockRoundTripper(ctrl)
	c := &validateHeaderTransport{
		Wrapped: rt,
		Allowed: map[string][]string{"client": {"browser", "mobile"}},
	}
	r := &http.Request{Header: http.Header{
		"client": {"mobile"},
	}}
	rt.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Status: "200 OK",
			StatusCode: http.StatusOK,
			Body: http.NoBody,
		},
		nil,
	).AnyTimes()
	got, err := c.RoundTrip(r)
	fmt.Println(got)
	assert.NoError(t, err)
}
