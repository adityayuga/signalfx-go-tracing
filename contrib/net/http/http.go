// Package http provides functions to trace the net/http package (https://golang.org/pkg/net/http).
package http // import "github.com/adityayuga/signalfx-go-tracing/contrib/net/http"

import (
	"net/http"

	"github.com/adityayuga/signalfx-go-tracing/contrib/internal/httputil"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
)

// ServeMux is an HTTP request multiplexer that traces all the incoming requests.
type ServeMux struct {
	*http.ServeMux
	cfg *config
}

// NewServeMux allocates and returns an http.ServeMux augmented with the
// global tracer.
func NewServeMux(opts ...Option) *ServeMux {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	return &ServeMux{
		ServeMux: http.NewServeMux(),
		cfg:      cfg,
	}
}

// ServeHTTP dispatches the request to the handler
// whose pattern most closely matches the request URL.
// We only need to rewrite this function to be able to trace
// all the incoming requests to the underlying multiplexer
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the resource associated to this request
	_, route := mux.Handler(r)
	opts := mux.cfg.spanOpts
	if mux.cfg.analyticsRate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, mux.cfg.analyticsRate))
	}
	httputil.TraceAndServe(mux.ServeMux, w, r, mux.cfg.serviceName, route, opts...)
}

// WrapHandler wraps an http.Handler with tracing using the given service and resource.
func WrapHandler(h http.Handler, service, resource string, opts ...Option) http.Handler {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		httputil.TraceAndServe(h, w, req, service, resource, cfg.spanOpts...)
	})
}
