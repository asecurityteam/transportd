package transportd

//go:generate mockgen -destination mock_backendregistry_test.go -package transportd github.com/asecurityteam/transportd/pkg BackendRegistry
//go:generate mockgen -destination mock_clientregistry_test.go -package transportd github.com/asecurityteam/transportd/pkg ClientRegistry
//go:generate mockgen -destination mock_backend_test.go -package transportd github.com/asecurityteam/transportd/pkg Backend
//go:generate mockgen -destination mock_source_test.go -package transportd github.com/asecurityteam/settings Source
//go:generate mockgen -destination mock_roundtripper_test.go -package transportd net/http RoundTripper
