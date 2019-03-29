package components

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/vincent-petithory/dataurl"
)

const (
	validToken   = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c` //nolint
	invalidToken = `NOTATOKEN`
)

func Test_asapValidateTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name             string
		r                *http.Request
		want             int
		wantErr          bool
		wantValidate     bool
		validateResponse error
	}{
		{
			name:             "missing bearer",
			r:                &http.Request{Header: http.Header{}},
			want:             http.StatusUnauthorized,
			wantErr:          false,
			wantValidate:     false,
			validateResponse: nil,
		},
		{
			name:             "invalid token",
			r:                &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + invalidToken}}},
			want:             http.StatusUnauthorized,
			wantErr:          false,
			wantValidate:     false,
			validateResponse: nil,
		},
		{
			name:             "validation failed",
			r:                &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + validToken}}},
			want:             http.StatusUnauthorized,
			wantErr:          false,
			wantValidate:     true,
			validateResponse: errors.New(""),
		},
		{
			name:             "validation passed",
			r:                &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + validToken}}},
			want:             http.StatusOK,
			wantErr:          false,
			wantValidate:     true,
			validateResponse: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			validator := NewMockValidator(ctrl)
			rt := NewMockRoundTripper(ctrl)
			var c = &asapValidateTransport{
				Wrapped:   rt,
				Validator: validator,
			}
			if tt.wantValidate {
				validator.EXPECT().Validate(gomock.Any()).Return(tt.validateResponse)
			}
			rt.EXPECT().RoundTrip(gomock.Any()).Return(
				&http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				},
				nil,
			).AnyTimes()
			var got, err = c.RoundTrip(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("asapValidateTransport.RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.StatusCode, tt.want) {
				t.Errorf("asapValidateTransport.RoundTrip() = %v, want %v", got.StatusCode, tt.want)
			}
		})
	}
}

const (
	aud     = "testAudience"
	iss     = "testIssuer"
	keyURL1 = "https://localhost"
	keyURL2 = "https://localhost"
)

func TestASAPValidateComponent_New(t *testing.T) {
	tests := []struct {
		name    string
		conf    *ASAPValidateConfig
		wantErr bool
	}{
		{
			name: "empty or missing issuers",
			conf: &ASAPValidateConfig{
				AllowedAudience: aud,
				KeyURLs:         []string{keyURL1, keyURL2},
			},
			wantErr: true,
		},
		{
			name: "empty or missing audience",
			conf: &ASAPValidateConfig{
				AllowedIssuers: []string{iss},
				KeyURLs:        []string{keyURL1, keyURL2},
			},
			wantErr: true,
		},
		{
			name: "empty or missing key urls",
			conf: &ASAPValidateConfig{
				AllowedIssuers:  []string{iss},
				AllowedAudience: aud,
			},
			wantErr: true,
		},
		{
			name: "success",
			conf: &ASAPValidateConfig{
				AllowedIssuers:  []string{iss},
				AllowedAudience: aud,
				KeyURLs:         []string{keyURL1, keyURL2},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ASAPValidateComponent{}
			_, err := m.New(context.Background(), tt.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ASAPValidateComponent.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

const (
	kid      = "testKid"
	tokenTTL = time.Hour
)

func TestASAPTokenComponent_New(t *testing.T) {
	pkBytes, _ := rsa.GenerateKey(rand.Reader, 2048)
	pkBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pkBytes),
	}
	pk := pem.EncodeToMemory(pkBlock)
	dataURIPK := dataurl.EncodeBytes(pk)
	tests := []struct {
		name    string
		conf    *ASAPTokenConfig
		wantErr bool
	}{
		{
			name: "missing or empty PK",
			conf: &ASAPTokenConfig{
				KID:       kid,
				TTL:       tokenTTL,
				Issuer:    iss,
				Audiences: []string{aud},
			},
			wantErr: true,
		},
		{
			name: "missing or empty KID",
			conf: &ASAPTokenConfig{
				PrivateKey: string(pk),
				TTL:        tokenTTL,
				Issuer:     iss,
				Audiences:  []string{aud},
			},
			wantErr: true,
		},
		{
			name: "missing or empty issuer",
			conf: &ASAPTokenConfig{
				PrivateKey: string(pk),
				KID:        kid,
				TTL:        tokenTTL,
				Audiences:  []string{aud},
			},
			wantErr: true,
		},
		{
			name: "missing or empty audiences",
			conf: &ASAPTokenConfig{
				PrivateKey: string(pk),
				KID:        kid,
				TTL:        tokenTTL,
				Issuer:     iss,
			},
			wantErr: true,
		},
		{
			name: "success",
			conf: &ASAPTokenConfig{
				PrivateKey: string(pk),
				KID:        kid,
				TTL:        tokenTTL,
				Issuer:     iss,
				Audiences:  []string{aud},
			},
			wantErr: false,
		},
		{
			name: "success-data-uri",
			conf: &ASAPTokenConfig{
				PrivateKey: dataURIPK,
				KID:        kid,
				TTL:        tokenTTL,
				Issuer:     iss,
				Audiences:  []string{aud},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ASAPTokenComponent{}
			_, err := a.New(context.Background(), tt.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ASAPTokenComponent.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
