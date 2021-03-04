package transportd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/asecurityteam/settings"
	"github.com/asecurityteam/transport"
)

const (
	enabledSetting = "enabled"
	backendSetting = "backend"
)

// ClientFactory exposes a Client constructor method that is bound
// to a given sregistry and set of plugin components.
type ClientFactory struct {
	Bases      BackendRegistry
	Components []NewComponent
}

// New generates a decorated http.RoundTripper for the given path and method.
// This method is used to handle the per-operation x-transportd blocks.
func (f *ClientFactory) New(ctx context.Context, s settings.Source, path string, method string) (http.RoundTripper, error) {
	componentsEnabled := settings.NewStringSliceSetting(enabledSetting, "", []string{})
	backendSelected := settings.NewStringSetting(backendSetting, "", "")
	enabledG := &settings.SettingGroup{
		NameValue:     ExtensionKey,
		SettingValues: []settings.Setting{componentsEnabled, backendSelected},
	}
	err := settings.LoadGroups(ctx, s, []settings.Group{enabledG})
	if err != nil {
		return nil, err
	}
	enabled := *componentsEnabled.StringSliceValue
	backend := *backendSelected.StringValue

	base := f.Bases.Load(ctx, backend)
	if base == nil {
		return nil, fmt.Errorf("backend %s not found for %s.%s", backend, path, method)
	}

	loadedComponents := make([]interface{}, len(enabled))
	for _, c := range f.Components {
		loadedComponent, err := c(ctx, backend, path, method)
		if err != nil {
			return nil, fmt.Errorf("failed to load component %v: %s", c, err.Error())
		}
		g, err := settings.GroupFromComponent(loadedComponent)
		if err != nil {
			return nil, fmt.Errorf("failed to load component %v: %s", c, err.Error())
		}
		for offset, en := range enabled {
			if strings.EqualFold(en, g.Name()) {
				loadedComponents[offset] = loadedComponent
			}
		}
	}
	for offset, en := range enabled {
		if loadedComponents[offset] == nil {
			return nil, fmt.Errorf("enabled component %s is not installed", en)
		}
	}
	prefixSource := &settings.PrefixSource{
		Source: s,
		Prefix: []string{ExtensionKey},
	}
	chain := make(transport.Chain, 0, len(enabled)+1)
	chain = append(chain, func(w http.RoundTripper) http.RoundTripper {
		return &hostRewrite{
			Wrapped: w,
			Host:    base.Host().Host,
			Scheme:  base.Host().Scheme,
		}
	})
	for offset, c := range loadedComponents {
		cD := new(func(http.RoundTripper) http.RoundTripper)
		err := settings.NewComponent(ctx, prefixSource, c, cD)
		if err != nil {
			return nil, fmt.Errorf("failed to load component %s: %s", enabled[offset], err.Error())
		}
		chain = append(chain, *cD)
	}
	return chain.Apply(base), nil
}
