// +build integration

package inttest

import (
	"context"
	"testing"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/asecurityteam/transportd/pkg/components"
	packr "github.com/gobuffalo/packr/v2"
)

type tc struct {
	Name    string
	Spec    []byte
	WantErr bool
}

func TestNewService(t *testing.T) {
	data := packr.New("data", "./specs")

	tcs := []tc{
		{
			Name:    "success",
			Spec:    data.Bytes("complete.yaml"),
			WantErr: false,
		},
		{
			Name:    "missing runtime",
			Spec:    data.Bytes("missingruntime.yaml"),
			WantErr: true,
		},
		{
			Name:    "passthrough enabled",
			Spec:    data.Bytes("passthroughenabled.yaml"),
			WantErr: false,
		},
	}

	for _, tt := range tcs {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := transportd.New(
				context.Background(),
				[]byte(tt.Spec),
				components.Defaults...,
			)
			if (err != nil) != tt.WantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}
