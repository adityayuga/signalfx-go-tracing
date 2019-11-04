package mux

import (
	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/internal/globalconfig"
)

type routerConfig struct {
	serviceName   string
	spanOpts      []ddtrace.StartSpanOption // additional span options to be applied
	analyticsRate float64
}

// RouterOption represents an option that can be passed to NewRouter.
type RouterOption func(*routerConfig)

func defaults(cfg *routerConfig) {
	cfg.analyticsRate = globalconfig.AnalyticsRate()
	cfg.serviceName = "mux.router"
}

// WithServiceName sets the given service name for the router.
func WithServiceName(name string) RouterOption {
	return func(cfg *routerConfig) {
		cfg.serviceName = name
	}
}

// WithSpanOptions applies the given set of options to the spans started
// by the router.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) RouterOption {
	return func(cfg *routerConfig) {
		cfg.spanOpts = opts
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) RouterOption {
	if on {
		return WithAnalyticsRate(1.0)
	}
	return WithAnalyticsRate(0.0)
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) RouterOption {
	return func(cfg *routerConfig) {
		cfg.analyticsRate = rate
	}
}
