package transportd

import (
	"encoding/json"

	"github.com/asecurityteam/settings"
)

const (
	// ExtensionKey is used to identify OpenAPI extension blocks
	// that are relevant to this project.
	ExtensionKey = "x-transportd"
)

// SourceFromExtension generates a settings.Source from any given
// OpenAPI extension block. The resulting Source tree has a root
// that matches the ExtensionKey.
func SourceFromExtension(s []byte) (settings.Source, error) {
	raw := make(map[string]interface{})
	err := json.Unmarshal(s, &raw)
	if err != nil {
		return nil, err
	}
	return settings.NewMapSource(map[string]interface{}{ExtensionKey: raw}), nil
}
