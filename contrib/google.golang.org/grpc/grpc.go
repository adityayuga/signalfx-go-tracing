//go:generate protoc -I . fixtures_test.proto --go_out=plugins=grpc:.

// Package grpc provides functions to trace the google.golang.org/grpc package v1.2.
package grpc // import "github.com/adityayuga/signalfx-go-tracing/contrib/google.golang.org/grpc"

import (
	"io"

	"github.com/adityayuga/signalfx-go-tracing/contrib/google.golang.org/internal/grpcutil"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"

	context "golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func startSpanFromContext(
	ctx context.Context, method, operation, service string, rate float64,
) (ddtrace.Span, context.Context) {
	opts := []ddtrace.StartSpanOption{
		tracer.ServiceName(service),
		tracer.ResourceName(method),
		tracer.Tag(tagMethod, method),
		tracer.SpanType(ext.AppTypeRPC),
	}
	if rate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
	}
	md, _ := metadata.FromIncomingContext(ctx) // nil is ok
	if sctx, err := tracer.Extract(grpcutil.MDCarrier(md)); err == nil {
		opts = append(opts, tracer.ChildOf(sctx))
	}
	return tracer.StartSpanFromContext(ctx, operation, opts...)
}

// finishWithError applies finish option and a tag with gRPC status code, disregarding OK, EOF and Canceled errors.
func finishWithError(span ddtrace.Span, err error, cfg *config) {
	if err == io.EOF || err == context.Canceled {
		err = nil
	}
	errcode := status.Code(err)
	if errcode == codes.OK || cfg.nonErrorCodes[errcode] {
		err = nil
	}
	span.SetTag(tagCode, errcode.String())
	finishOptions := []tracer.FinishOption{
		tracer.WithError(err),
	}
	if cfg.noDebugStack {
		finishOptions = append(finishOptions, tracer.NoDebugStack())
	}
	span.FinishWithOptionsExt(finishOptions...)
}
