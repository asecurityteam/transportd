package transportd

import (
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
)

type ctxKey string

var (
	routeCtxKey  = ctxKey("__transportd_route")
	paramsCtxKey = ctxKey("__transportd_params")
)

// RouteFromContext fetches the active OpenAPI route.
func RouteFromContext(ctx context.Context) *routers.Route {
	return ctx.Value(routeCtxKey).(*routers.Route)
}

// PathParamsFromContext fetches the matching URL params.
func PathParamsFromContext(ctx context.Context) map[string]string {
	return ctx.Value(paramsCtxKey).(map[string]string)
}

// RouteToContext inserts the active OpenAPI route.
func RouteToContext(ctx context.Context, route *routers.Route) context.Context {
	return context.WithValue(ctx, routeCtxKey, route)
}

// PathParamsToContext inserts the matching URL params.
func PathParamsToContext(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, paramsCtxKey, params)
}

// ClientTransport maps incoming requests to a configured client.
type ClientTransport struct {
	Registry ClientRegistry
	Router   routers.Router
}

// RoundTrip performs a client lookup and uses the result to execute the
// request.
//
// If a client is not found then a NotFound response it returned unless there
// is a route for the path "unknown" and the method "unknown".
//
// If a client is found then the Route is injected into the request context
// for later use.
func (r *ClientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	route, pathParams, err := r.Router.FindRoute(req)
	if err != nil {
		defaultTr := r.Registry.Load(req.Context(), unknownKey, unknownKey)
		if defaultTr == nil {
			return newError(http.StatusNotFound, err.Error()), nil
		}
		route = &routers.Route{
			Method: req.Method,
			Path:   unknownKey,
			Operation: &openapi3.Operation{
				OperationID: unknownKey,
			},
		}
		req = req.WithContext(RouteToContext(req.Context(), route))
		req = req.WithContext(PathParamsToContext(req.Context(), make(map[string]string)))
		return defaultTr.RoundTrip(req)
	}
	req = req.WithContext(RouteToContext(req.Context(), route))
	req = req.WithContext(PathParamsToContext(req.Context(), pathParams))
	return r.Registry.Load(req.Context(), route.Path, route.Method).RoundTrip(req)
}
