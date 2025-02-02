package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/adityayuga/signalfx-go-tracing/contrib/database/sql/internal"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
)

var _ driver.Driver = (*tracedDriver)(nil)

// tracedDriver wraps an inner sql driver with tracing. It implements the (database/sql).driver.Driver interface.
type tracedDriver struct {
	driver.Driver
	driverName string
	config     *registerConfig
}

// Open returns a tracedConn so that we can pass all the info we get from the DSN
// all along the tracing
func (d *tracedDriver) Open(dsn string) (c driver.Conn, err error) {
	var (
		meta map[string]string
		conn driver.Conn
	)
	meta, err = internal.ParseDSN(d.driverName, dsn)
	if err != nil {
		return nil, err
	}
	conn, err = d.Driver.Open(dsn)
	if err != nil {
		return nil, err
	}
	tp := &traceParams{
		driverName: d.driverName,
		config:     d.config,
		meta:       meta,
	}
	return &tracedConn{conn, tp}, err
}

// traceParams stores all information relative to the tracing
type traceParams struct {
	config     *registerConfig
	driverName string
	resource   string
	meta       map[string]string
}

// tryTrace will create a span using the given arguments, but will act as a no-op when err is driver.ErrSkip.
func (tp *traceParams) tryTrace(ctx context.Context, resource string, query string, startTime time.Time, err error) {
	if err == driver.ErrSkip {
		// Not a user error: driver is telling sql package that an
		// optional interface method is not implemented. There is
		// nothing to trace here.
		// See: https://github.com/DataDog/dd-trace-go/issues/270
		return
	}
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeSQL),
		tracer.ServiceName(tp.config.serviceName),
		tracer.StartTime(startTime),
	}
	if rate := tp.config.analyticsRate; rate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
	}
	span, _ := tracer.StartSpanFromContext(ctx, resource, opts...)
	if query != "" {
		resource = query
		span.SetTag(ext.DBStatement, query)
	}
	span.SetTag(ext.ResourceName, resource)
	for k, v := range tp.meta {
		span.SetTag(k, v)
	}
	span.SetTag(ext.DBType, tp.driverName)
	span.FinishWithOptionsExt(tracer.WithError(err))
}

// tracedDriverName returns the name of the traced version for the given driver name.
func tracedDriverName(name string) string { return name + ".traced" }

// driverExists returns true if the given driver name has already been registered.
func driverExists(name string) bool {
	for _, v := range sql.Drivers() {
		if name == v {
			return true
		}
	}
	return false
}
