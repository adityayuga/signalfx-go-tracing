// Package graphql provides functions to trace the graph-gophers/graphql-go package (https://github.com/graph-gophers/graphql-go).
//
// We use the tracing mechanism available in the
// https://godoc.org/github.com/graph-gophers/graphql-go/trace subpackage.
// Create a new Tracer with `NewTracer` and pass it as an additional option to
// `MustParseSchema`.
package graphql // import "github.com/adityayuga/signalfx-go-tracing/contrib/graph-gophers/graphql-go"

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/trace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
)

const (
	tagGraphqlField = "graphql.field"
	tagGraphqlQuery = "graphql.query"
	tagGraphqlType  = "graphql.type"
)

// A Tracer implements the graphql-go/trace.Tracer interface by sending traces
// to the Datadog tracer.
type Tracer struct {
	cfg *config
}

var _ trace.Tracer = (*Tracer)(nil)

// TraceQuery traces a GraphQL query.
func (t *Tracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	opts := []ddtrace.StartSpanOption{
		tracer.ServiceName(t.cfg.serviceName),
		tracer.Tag(tagGraphqlQuery, queryString),
	}
	if t.cfg.analyticsRate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, t.cfg.analyticsRate))
	}
	span, ctx := tracer.StartSpanFromContext(ctx, "graphql.request", opts...)

	return ctx, func(errs []*errors.QueryError) {
		var err error
		switch n := len(errs); n {
		case 0:
			// err = nil
		case 1:
			err = errs[0]
		default:
			err = fmt.Errorf("%s (and %d more errors)", errs[0], n-1)
		}
		span.FinishWithOptionsExt(tracer.WithError(err))
	}
}

// TraceField traces a GraphQL field access.
func (t *Tracer) TraceField(ctx context.Context, label string, typeName string, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	opts := []ddtrace.StartSpanOption{
		tracer.ServiceName(t.cfg.serviceName),
		tracer.Tag(tagGraphqlField, fieldName),
		tracer.Tag(tagGraphqlType, typeName),
	}
	if t.cfg.analyticsRate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, t.cfg.analyticsRate))
	}
	span, ctx := tracer.StartSpanFromContext(ctx, "graphql.field", opts...)

	return ctx, func(err *errors.QueryError) {
		// must explicitly check for nil, see issue golang/go#22729
		if err != nil {
			span.FinishWithOptionsExt(tracer.WithError(err))
		} else {
			span.Finish()
		}
	}
}

// NewTracer creates a new Tracer.
func NewTracer(opts ...Option) trace.Tracer {
	cfg := new(config)
	defaults(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	return &Tracer{
		cfg: cfg,
	}
}
