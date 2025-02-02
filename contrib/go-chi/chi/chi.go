// Package chi provides tracing functions for tracing the go-chi/chi package (https://github.com/go-chi/chi).
package chi // import "github.com/adityayuga/signalfx-go-tracing/contrib/go-chi/chi"

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
)

// Middleware returns middleware that will trace incoming requests.
func Middleware(opts ...Option) func(next http.Handler) http.Handler {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opts := []ddtrace.StartSpanOption{
				tracer.SpanType(ext.SpanTypeWeb),
				tracer.ServiceName(cfg.serviceName),
				tracer.Tag(ext.HTTPMethod, r.Method),
				tracer.Tag(ext.HTTPURL, r.URL.Path),
			}
			if cfg.analyticsRate > 0 {
				opts = append(opts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
			}
			if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(r.Header)); err == nil {
				opts = append(opts, tracer.ChildOf(spanctx))
			}
			opts = append(opts, cfg.spanOpts...)
			span, ctx := tracer.StartSpanFromContext(r.Context(), "http.request", opts...)
			defer span.Finish()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// pass the span through the request context and serve the request to the next middleware
			next.ServeHTTP(ww, r.WithContext(ctx))

			// set the resource name as we get it only once the handler is executed
			resourceName := chi.RouteContext(r.Context()).RoutePattern()
			if resourceName == "" {
				resourceName = "unknown"
			}
			resourceName = r.Method + " " + resourceName
			span.SetTag(ext.ResourceName, resourceName)

			// set the status code
			status := ww.Status()
			span.SetTag(ext.HTTPCode, strconv.Itoa(status))

			if status >= 500 && status < 600 {
				// mark 5xx server error
				span.SetTag(ext.Error, fmt.Errorf("%d: %s", status, http.StatusText(status)))
			}
		})
	}
}
