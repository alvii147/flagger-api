package httputils

import (
	"fmt"
	"net/http"
)

// MiddlewareFunc takes in a HandlerFunc and returns a HandlerFunc.
// Middleware is used to add a processing step to handlers.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Router represents a handler that can register handlers under specific HTTP methods.
type Router interface {
	GET(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	POST(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	PUT(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	DELETE(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// router implements Router.
type router struct {
	*http.ServeMux
}

// NewRouter creates and returns a new router.
func NewRouter(mux *http.ServeMux) *router {
	return &router{
		ServeMux: mux,
	}
}

// handleWithMiddleware wraps a handler with a set of middleware
// registers it for a given pattern and HTTP method.
func (r *router) handleWithMiddleware(
	method string,
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	wrappedHandler := handler
	for _, mw := range middleware {
		wrappedHandler = mw(wrappedHandler)
	}

	r.ServeMux.Handle(
		fmt.Sprintf("%s %s", method, pattern),
		ResponseWriterMiddleware(wrappedHandler),
	)
}

// GET wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP GET method.
func (r *router) GET(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.handleWithMiddleware(http.MethodGet, pattern, handler, middleware...)
}

// POST wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP POST method.
func (r *router) POST(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.handleWithMiddleware(http.MethodPost, pattern, handler, middleware...)
}

// PUT wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP PUT method.
func (r *router) PUT(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.handleWithMiddleware(http.MethodPut, pattern, handler, middleware...)
}

// DELETE wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP DELETE method.
func (r *router) DELETE(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.handleWithMiddleware(http.MethodDelete, pattern, handler, middleware...)
}
