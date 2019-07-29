package transportd

import (
	"fmt"
	"net/http"
	"testing"
)

func TestEnforceRelativeLocation(t *testing.T) {
	tests := []struct {
		name         string
		resp         *http.Response
		wantErr      bool
		wantLocation string
	}{
		{
			name: "missing location",
			resp: &http.Response{
				Header: http.Header{},
			},
			wantErr:      false,
			wantLocation: "",
		},
		{
			name: "invalid location",
			resp: &http.Response{
				Header: http.Header{
					"Location": []string{"https://[XXX]:notport/path"},
				},
			},
			wantErr:      false,
			wantLocation: "https://[XXX]:notport/path",
		},
		{
			name: "relative location",
			resp: &http.Response{
				Header: http.Header{
					"Location": []string{"/path/to/api?q=1"},
				},
			},
			wantErr:      false,
			wantLocation: "/path/to/api?q=1",
		},
		{
			name: "relative location (trailing)",
			resp: &http.Response{
				Header: http.Header{
					"Location": []string{"/path/to/api/?q=1"},
				},
			},
			wantErr:      false,
			wantLocation: "/path/to/api/?q=1",
		},
		{
			name: "absolute location",
			resp: &http.Response{
				Header: http.Header{
					"Location": []string{"https://localhost/path/to/api?q=1"},
				},
			},
			wantErr:      false,
			wantLocation: "/path/to/api?q=1",
		},
		{
			name: "absolute location (trailing)",
			resp: &http.Response{
				Header: http.Header{
					"Location": []string{"https://localhost/path/to/api/?q=1"},
				},
			},
			wantErr:      false,
			wantLocation: "/path/to/api/?q=1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnforceRelativeLocation(tt.resp); (err != nil) != tt.wantErr {
				t.Errorf("EnforceRelativeLocation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.wantLocation != tt.resp.Header.Get("Location") {
				t.Errorf("EnforceRelativeLocation() Location = %v, wantErr %v", tt.resp.Header.Get("Location"), tt.wantLocation)
			}
		})
	}
}

func TestMultiResponseModifier_ModifyResponse(t *testing.T) {
	tests := []struct {
		name    string
		mrs     MultiResponseModifier
		resp    *http.Response
		wantErr bool
	}{
		{
			name:    "empty",
			mrs:     MultiResponseModifier{},
			resp:    &http.Response{},
			wantErr: false,
		},
		{
			name: "success",
			mrs: MultiResponseModifier{
				func(*http.Response) error {
					return nil
				},
				func(*http.Response) error {
					return nil
				},
			},
			resp:    &http.Response{},
			wantErr: false,
		},
		{
			name: "error",
			mrs: MultiResponseModifier{
				func(*http.Response) error {
					return nil
				},
				func(*http.Response) error {
					return fmt.Errorf("")
				},
			},
			resp:    &http.Response{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mrs.ModifyResponse(tt.resp); (err != nil) != tt.wantErr {
				t.Errorf("MultiResponseModifier.ModifyResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
