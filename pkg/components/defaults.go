package components

import (
	transportd "github.com/asecurityteam/transportd/pkg"
)

var (
	// Defaults is the same set of components that is installed in
	// the project main.go and can be used by custom builds to stay in sync
	// with the latest additions.
	Defaults = []transportd.NewComponent{
		Metrics,
		AccessLog,
		ASAPValidate,
		Timeout,
		Hedging,
		Retry,
		ASAPToken,
		RequestValidation,
		ResponseValidation,
		Strip,
		Header,
		BasicAuth,
		Auth,
	}
)
