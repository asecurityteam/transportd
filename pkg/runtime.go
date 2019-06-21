package transportd

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
)

const (
	// RuntimeExtensionKey is used to identify the runtime configuration
	// extension block at the top level of an OpenAPI specification.
	RuntimeExtensionKey = "x-runtime"
)

// RuntimeSourceFromExtension is a one-off change from the SourceFromExtension
// method that handles the runhttp configuration block. This is needed
// because the runhttp component has a predefined root of "runtime"
// that we need to adapt the source to match.
func RuntimeSourceFromExtension(s []byte) (settings.Source, error) {
	rt := runhttp.NewComponent()
	grp, _ := settings.GroupFromComponent(rt)

	raw := make(map[string]interface{})
	err := json.Unmarshal(s, &raw)
	if err != nil {
		return nil, err
	}
	return settings.NewMapSource(map[string]interface{}{grp.Name(): raw}), nil
}

// NewRuntime generates a runhttp.Runtime instance that will host the
// given handler. This method is used to handle the top-level x-runtime
// block.
func NewRuntime(ctx context.Context, s settings.Source, h http.Handler) (*runhttp.Runtime, error) {
	rt := runhttp.NewComponent().WithHandler(h)
	rtD := new(runhttp.Runtime)
	err := settings.NewComponent(ctx, s, rt, rtD)
	return rtD, err
}
