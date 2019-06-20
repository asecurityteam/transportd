package transportd

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
)

type rtC struct {
	*runhttp.Config
}

func (*rtC) Name() string {
	return RuntimeExtensionKey
}

// Help outputs a formatted string to help with discovering available
// settings and options.
func Help(ctx context.Context, components ...NewComponent) (string, error) {
	var result bytes.Buffer

	rt := runhttp.NewComponent().Settings()
	rtG, _ := settings.Convert(&rtC{rt})
	_, _ = result.WriteString("The following top level extension must appear and configures the runtime:\n")
	_, _ = result.WriteString(settings.ExampleYamlGroups([]settings.Group{rtG}))
	_, _ = result.WriteString("\n")

	backendsInstalled := settings.NewStringSliceSetting(backendsSetting, "Available backends.", []string{"backendName"})
	host := settings.NewStringSetting(hostSetting, "Backend host URL.", "https://localhost")
	poolCount := settings.NewIntSetting(countSetting, "Number of connections pools. Only use >1 if HTTP/2", 1)
	poolTTL := settings.NewDurationSetting(ttlSetting, "Lifetime of a pool before refreshing.", time.Hour)
	pool := &settings.SettingGroup{
		NameValue:     poolSetting,
		SettingValues: []settings.Setting{poolCount, poolTTL},
	}
	backendsG := &settings.SettingGroup{
		NameValue:     ExtensionKey,
		SettingValues: []settings.Setting{backendsInstalled},
		GroupValues: []settings.Group{
			&settings.SettingGroup{
				NameValue:        "backendName",
				DescriptionValue: "Configuration for a single backend.",
				SettingValues:    []settings.Setting{host},
				GroupValues:      []settings.Group{pool},
			},
		},
	}
	_, _ = result.WriteString("The following top level extension must appear and configures service backends:\n")
	_, _ = result.WriteString(settings.ExampleYamlGroups([]settings.Group{backendsG}))
	_, _ = result.WriteString("\n")

	backendSelected := settings.NewStringSetting(backendSetting, "Backend target for this route.", "backendName")
	componentConfigs := make([]settings.Group, 0, len(components))
	names := make([]string, 0, len(components))
	for _, comp := range components {
		c, err := comp(ctx, "", "", "")
		if err != nil {
			return "", fmt.Errorf("failed to generate help: %s", err.Error())
		}
		g, err := settings.GroupFromComponent(c)
		if err != nil {
			return "", fmt.Errorf("failed to generate help: %s", err.Error())
		}
		componentConfigs = append(componentConfigs, g)
		names = append(names, g.Name())
	}
	componentsEnabled := settings.NewStringSliceSetting(enabledSetting, "Ordered list of components enabled for this route.", names)
	enabledG := &settings.SettingGroup{
		NameValue:     ExtensionKey,
		SettingValues: []settings.Setting{componentsEnabled, backendSelected},
		GroupValues:   componentConfigs,
	}
	_, _ = result.WriteString("The following per-route extension must appear and configures request behavior:\n")
	_, _ = result.WriteString(settings.ExampleYamlGroups([]settings.Group{enabledG}))

	return result.String(), nil
}
