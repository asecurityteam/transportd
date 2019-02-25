package transportd

import (
	"os"
	"regexp"
)

const (
	envPattern = `\${[^}]+}`
)

// EnvProcessor transforms documents by interpolating environment variables.
type EnvProcessor struct {
	pattern *regexp.Regexp
	osenv   func(string) string
}

// NewEnvProcessor prepares the EnvProcessor and returns it.
func NewEnvProcessor() *EnvProcessor {
	var p, _ = regexp.Compile(envPattern)
	return &EnvProcessor{pattern: p}
}

// Process replaces ${} values with values from the environment.
func (y *EnvProcessor) Process(source []byte) ([]byte, error) {
	osenv := y.osenv
	if osenv == nil {
		osenv = os.Getenv
	}
	return y.pattern.ReplaceAllFunc(source, func(match []byte) []byte {
		name := match[2 : len(match)-1] // strip ${}
		return []byte(osenv(string(name)))
	}), nil
}
