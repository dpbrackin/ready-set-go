// Pakage router implements simple http routing using [net/http].
package router

import (
	"net/http"
	"slices"
)

type Root struct {
	mux         *http.ServeMux
	middlewares []Middleware
	routes      map[string]http.Handler
	groups      []*RouteGroup
}

// RouteGroup groups related routes under a common prefix and use the same middlewares.
type RouteGroup struct {
	prefix      string
	middlewares []Middleware
	routes      map[string]http.Handler
}

type Middleware func(next http.Handler) http.Handler

func applyMiddlewares(f http.Handler, m []Middleware) http.Handler {
	for _, middleware := range slices.Backward(m) {
		f = middleware(f)
	}

	return f
}

func NewRootRouter() *Root {
	return &Root{
		middlewares: make([]Middleware, 0),
		groups:      make([]*RouteGroup, 0),
		routes:      make(map[string]http.Handler),
		mux:         http.NewServeMux(),
	}
}

// Group adds a new RouteGroup to the router.
func (router *Root) Group(prefix string) *RouteGroup {
	group := &RouteGroup{
		prefix:      prefix,
		middlewares: make([]Middleware, 0),
		routes:      make(map[string]http.Handler),
	}

	router.groups = append(router.groups, group)

	return group
}

// Use adds a middleware that is used for all routes in the router.
// Middlewares are applied in the same order `Use` is called.
func (router *Root) Use(middleware Middleware) {
	router.middlewares = append(router.middlewares, middleware)
}

// RouteFunc adds a route that is handled by a function
func (router *Root) RouteFunc(route string, f func(http.ResponseWriter, *http.Request)) {
	router.routes[route] = http.HandlerFunc(f)
}

// Use adds a middleware that is used for all routes in the router.
// Middlewares are applied in the same order `Use` is called.
func (group *RouteGroup) Use(middleware Middleware) {
	group.middlewares = append(group.middlewares, middleware)
}

// RouteFunc adds a route that is handled by a function to the group.
func (group *RouteGroup) RouteFunc(route string, f func(http.ResponseWriter, *http.Request)) {
	group.routes[route] = http.HandlerFunc(f)
}

func (router *Root) Mux() *http.ServeMux {
	for route, handler := range router.routes {
		handlerWithMiddlewares := applyMiddlewares(handler, router.middlewares)
		router.mux.Handle(route, handlerWithMiddlewares)
	}

	for _, group := range router.groups {
		middlewares := append(router.middlewares, group.middlewares...)

		for route, handler := range group.routes {
			handlerWithMiddlewares := applyMiddlewares(handler, middlewares)

			method, host, path := parsePattern(route)

			router.mux.Handle(method+" "+host+group.prefix+path, handlerWithMiddlewares)
		}
	}

	return router.mux
}
