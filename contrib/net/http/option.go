package http

import (
	"net/http"

	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/internal/globalconfig"
)

type config struct {
	serviceName   string
	analyticsRate float64
	spanOpts      []ddtrace.StartSpanOption
}

// MuxOption has been deprecated in favor of Option.
type MuxOption func(*config)

// Option represents an option that can be passed to NewServeMux or WrapHandler.
type Option = MuxOption

func defaults(cfg *config) {
	cfg.analyticsRate = globalconfig.AnalyticsRate()
	cfg.serviceName = "http.router"
}

// WithServiceName sets the given service name for the returned ServeMux.
func WithServiceName(name string) MuxOption {
	return func(cfg *config) {
		cfg.serviceName = name
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) MuxOption {
	if on {
		return WithAnalyticsRate(1.0)
	}
	return WithAnalyticsRate(0.0)
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) MuxOption {
	return func(cfg *config) {
		cfg.analyticsRate = rate
	}
}

// WithSpanOptions defines a set of additional ddtrace.StartSpanOption to be added
// to spans started by the integration.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) Option {
	return func(cfg *config) {
		cfg.spanOpts = append(cfg.spanOpts, opts...)
	}
}

// A RoundTripperBeforeFunc can be used to modify a span before an http
// RoundTrip is made.
type RoundTripperBeforeFunc func(*http.Request, ddtrace.Span)

// A RoundTripperAfterFunc can be used to modify a span after an http
// RoundTrip is made. It is possible for the http Response to be nil.
type RoundTripperAfterFunc func(*http.Response, ddtrace.Span)

type roundTripperConfig struct {
	before        RoundTripperBeforeFunc
	after         RoundTripperAfterFunc
	analyticsRate float64
}

func newRoundTripperConfig() *roundTripperConfig {
	return &roundTripperConfig{
		analyticsRate: globalconfig.AnalyticsRate(),
	}
}

// A RoundTripperOption represents an option that can be passed to
// WrapRoundTripper.
type RoundTripperOption func(*roundTripperConfig)

// WithBefore adds a RoundTripperBeforeFunc to the RoundTripper
// config.
func WithBefore(f RoundTripperBeforeFunc) RoundTripperOption {
	return func(cfg *roundTripperConfig) {
		cfg.before = f
	}
}

// WithAfter adds a RoundTripperAfterFunc to the RoundTripper
// config.
func WithAfter(f RoundTripperAfterFunc) RoundTripperOption {
	return func(cfg *roundTripperConfig) {
		cfg.after = f
	}
}

// RTWithAnalytics enables Trace Analytics for all started spans.
func RTWithAnalytics(on bool) RoundTripperOption {
	if on {
		return RTWithAnalyticsRate(1.0)
	}
	return RTWithAnalyticsRate(0.0)
}

// RTWithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func RTWithAnalyticsRate(rate float64) RoundTripperOption {
	return func(cfg *roundTripperConfig) {
		cfg.analyticsRate = rate
	}
}
