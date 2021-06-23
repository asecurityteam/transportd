// +build integration

package inttest

import (
	"context"
	"embed"
	"path"
	"testing"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/asecurityteam/transportd/pkg/components"
)

type tc struct {
	Name    string
	File    string
	WantErr bool
}

//go:embed specs
var specs embed.FS

func TestNewService(t *testing.T) {
	folder := "specs"

	tcs := []tc{
		{
			Name:    "success",
			File:    "complete.yaml",
			WantErr: false,
		},
		{
			Name:    "missing runtime",
			File:    "missingruntime.yaml",
			WantErr: true,
		},
		{
			Name:    "passthrough enabled",
			File:    "passthroughenabled.yaml",
			WantErr: false,
		},
	}

	for _, tt := range tcs {
		t.Run(tt.Name, func(t *testing.T) {
			filePath := path.Join(folder, tt.File)
			fileInput, err := specs.ReadFile(filePath)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.WantErr)
			}
			_, err = transportd.New(
				context.Background(),
				fileInput,
				components.Defaults...,
			)
			if (err != nil) != tt.WantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}
