package components

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestBasicAuthComponent_New(t *testing.T) {
	user := "user"
	pass := "pass"
	tests := []struct {
		name    string
		conf    *BasicAuthConfig
		wantErr bool
	}{
		{
			name: "missing or empty username",
			conf: &BasicAuthConfig{
				Password: pass,
			},
			wantErr: true,
		},
		{
			name: "missing or empty password",
			conf: &BasicAuthConfig{
				Username: user,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BasicAuthComponent{}
			_, err := b.New(context.Background(), tt.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("BasicAuthComponent.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestBasicAuthTransport_RoundTrip(t *testing.T) {
	user := "user"
	pass := "pass"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rt := NewMockRoundTripper(ctrl)
	c := &basicAuthTransport{
		Wrapped:  rt,
		Username: user,
		Password: pass,
	}

	r := &http.Request{Header: http.Header{}}
	rt.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       http.NoBody,
		},
		nil,
	)

	_, _ = c.RoundTrip(r)
	u, p, ok := r.BasicAuth()
	if !ok {
		t.Errorf("basicAuthTransport.RoundTrip() did not set basic auth headers")
		return
	}
	if user != u || pass != p {
		t.Errorf("basicAuthTransport.RoundTrip() user = %s pass = %s, want %s %s", user, pass, u, p)
	}
}
