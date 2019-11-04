package tracer

import (
	"context"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Not needed with delegating to opentracing.
//func TestContextWithSpan(t *testing.T) {
//	want := &span{SpanID: 123}
//	ctx := ContextWithSpan(context.Background(), want)
//	got, ok := ctx.Value(activeSpanKey).(*span)
//	assert := assert.New(t)
//	assert.True(ok)
//	assert.Equal(got, want)
//}

func TestSpanFromContext(t *testing.T) {
	t.Run("regular", func(t *testing.T) {
		assert := assert.New(t)
		want := &span{SpanID: 123}
		ctx := ContextWithSpan(context.Background(), want)
		got, ok := SpanFromContext(ctx)
		assert.True(ok)
		assert.Equal(got, want)
	})
	t.Run("no-op", func(t *testing.T) {
		assert := assert.New(t)
		span, ok := SpanFromContext(context.Background())
		assert.False(ok)
		_, ok = span.(*ddtrace.NoopSpan)
		assert.True(ok)
		span, ok = SpanFromContext(nil)
		assert.False(ok)
		_, ok = span.(*ddtrace.NoopSpan)
		assert.True(ok)
	})
}

func TestStartSpanFromContext(t *testing.T) {
	_, _, stop := startTestTracer()
	defer stop()

	parent := &span{context: &spanContext{spanID: 123, traceID: 456}}
	parent2 := &span{context: &spanContext{spanID: 789, traceID: 456}}
	pctx := ContextWithSpan(context.Background(), parent)
	child, ctx := StartSpanFromContext(
		pctx,
		"http.request",
		ServiceName("gin"),
		ResourceName("/"),
		ChildOf(parent2.Context()), // we do this to assert that the span in pctx takes priority.
	)
	assert := assert.New(t)

	got, ok := child.(*span)
	assert.True(ok)
	gotctx, ok := SpanFromContext(ctx)
	assert.True(ok)
	assert.Equal(gotctx, got)
	_, ok = gotctx.(*ddtrace.NoopSpan)
	assert.False(ok)

	assert.Equal(uint64(456), got.TraceID)
	assert.Equal(uint64(123), got.ParentID)
	assert.Equal("http.request", got.Name)
	assert.Equal("gin", got.Service)
	assert.Equal("/", got.Resource)
}
