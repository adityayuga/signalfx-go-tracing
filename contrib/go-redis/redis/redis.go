// Package redis provides tracing functions for tracing the go-redis/redis package (https://github.com/go-redis/redis).
package redis

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"

	"github.com/go-redis/redis"
)

// Client is used to trace requests to a redis server.
type Client struct {
	*redis.Client
	*params

	mu  sync.RWMutex // guards ctx
	ctx context.Context
}

var _ redis.Cmdable = (*Client)(nil)

// Pipeliner is used to trace pipelines executed on a Redis server.
type Pipeliner struct {
	redis.Pipeliner
	*params

	ctx context.Context
}

var _ redis.Pipeliner = (*Pipeliner)(nil)

// params holds the tracer and a set of parameters which are recorded with every trace.
type params struct {
	host   string
	port   string
	db     string
	config *clientConfig
}

// NewClient returns a new Client that is traced with the default tracer under
// the service name "redis".
func NewClient(opt *redis.Options, opts ...ClientOption) *Client {
	return WrapClient(redis.NewClient(opt), opts...)
}

// WrapClient wraps a given redis.Client with a tracer under the given service name.
func WrapClient(c *redis.Client, opts ...ClientOption) *Client {
	cfg := new(clientConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	opt := c.Options()
	host, port, err := net.SplitHostPort(opt.Addr)
	if err != nil {
		host = opt.Addr
		port = "6379"
	}
	params := &params{
		host:   host,
		port:   port,
		db:     strconv.Itoa(opt.DB),
		config: cfg,
	}
	tc := &Client{Client: c, params: params}
	tc.Client.WrapProcess(createWrapperFromClient(tc))
	return tc
}

// Pipeline creates a Pipeline from a Client
func (c *Client) Pipeline() redis.Pipeliner {
	c.mu.RLock()
	ctx := c.ctx
	c.mu.RUnlock()
	return &Pipeliner{c.Client.Pipeline(), c.params, ctx}
}

// ExecWithContext calls Pipeline.Exec(). It ensures that the resulting Redis calls
// are traced, and that emitted spans are children of the given Context.
func (c *Pipeliner) ExecWithContext(ctx context.Context) ([]redis.Cmder, error) {
	return c.execWithContext(ctx)
}

// Exec calls Pipeline.Exec() ensuring that the resulting Redis calls are traced.
func (c *Pipeliner) Exec() ([]redis.Cmder, error) {
	return c.execWithContext(c.ctx)
}

func (c *Pipeliner) execWithContext(ctx context.Context) ([]redis.Cmder, error) {
	p := c.params
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeRedis),
		tracer.ServiceName(p.config.serviceName),
		tracer.ResourceName("redis"),
		tracer.Tag(ext.TargetHost, p.host),
		tracer.Tag(ext.TargetPort, p.port),
		tracer.Tag("out.db", p.db),
	}
	if rate := p.config.analyticsRate; rate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
	}
	span, _ := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
	cmds, err := c.Pipeliner.Exec()
	span.SetTag(ext.ResourceName, commandsToString(cmds))
	span.SetTag("redis.pipeline_length", strconv.Itoa(len(cmds)))
	var finishOpts []ddtrace.FinishOption
	if err != redis.Nil {
		finishOpts = append(finishOpts, tracer.WithError(err))
	}
	span.FinishWithOptionsExt(finishOpts...)

	return cmds, err
}

// commandsToString returns a string representation of a slice of redis Commands, separated by newlines.
func commandsToString(cmds []redis.Cmder) string {
	var b bytes.Buffer
	for _, cmd := range cmds {
		b.WriteString(cmderToString(cmd))
		b.WriteString("\n")
	}
	return b.String()
}

// WithContext sets a context on a Client. Use it to ensure that emitted spans have the correct parent.
func (c *Client) WithContext(ctx context.Context) *Client {
	c.mu.Lock()
	c.ctx = ctx
	c.mu.Unlock()
	return c
}

// Context returns the active context in the client.
func (c *Client) Context() context.Context {
	c.mu.RLock()
	ctx := c.ctx
	c.mu.RUnlock()
	return ctx
}

// createWrapperFromClient returns a new createWrapper function which wraps the processor with tracing
// information obtained from the provided Client. To understand this functionality better see the
// documentation for the github.com/go-redis/redis.(*baseClient).WrapProcess function.
func createWrapperFromClient(tc *Client) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			tc.mu.RLock()
			ctx := tc.ctx
			tc.mu.RUnlock()
			raw := cmderToString(cmd)
			parts := strings.Split(raw, " ")
			length := len(parts) - 1
			p := tc.params
			opts := []ddtrace.StartSpanOption{
				tracer.SpanType(ext.SpanTypeRedis),
				tracer.ServiceName(p.config.serviceName),
				tracer.ResourceName(parts[0]),
				tracer.Tag(ext.TargetHost, p.host),
				tracer.Tag(ext.TargetPort, p.port),
				tracer.Tag("out.db", p.db),
				tracer.Tag("redis.raw_command", raw),
				tracer.Tag("redis.args_length", strconv.Itoa(length)),
			}
			if rate := p.config.analyticsRate; rate > 0 {
				opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
			}
			span, _ := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
			err := oldProcess(cmd)
			var finishOpts []ddtrace.FinishOption
			if err != redis.Nil {
				finishOpts = append(finishOpts, tracer.WithError(err))
			}
			span.FinishWithOptionsExt(finishOpts...)
			return err
		}
	}
}

func cmderToString(cmd redis.Cmder) string {
	// We want to support multiple versions of the go-redis library. In
	// older versions Cmder implements the Stringer interface, while in
	// newer versions that was removed, and this String method which
	// sometimes returns an error is used instead. By doing a type assertion
	// we can support both versions.
	switch v := cmd.(type) {
	case fmt.Stringer:
		return v.String()
	case interface{ String() (string, error) }:
		str, err := v.String()
		if err == nil {
			return str
		}
	}
	args := cmd.Args()
	if len(args) == 0 {
		return ""
	}
	if str, ok := args[0].(string); ok {
		return str
	}
	return ""
}
