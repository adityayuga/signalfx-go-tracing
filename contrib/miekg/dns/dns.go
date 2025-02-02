package dns

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/miekg/dns"

	"github.com/adityayuga/signalfx-go-tracing/ddtrace"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
)

// ListenAndServe calls dns.ListenAndServe with a wrapped Handler.
func ListenAndServe(addr string, network string, handler dns.Handler) error {
	return dns.ListenAndServe(addr, network, WrapHandler(handler))
}

// ListenAndServeTLS calls dns.ListenAndServeTLS with a wrapped Handler.
func ListenAndServeTLS(addr, certFile, keyFile string, handler dns.Handler) error {
	return dns.ListenAndServeTLS(addr, certFile, keyFile, WrapHandler(handler))
}

// A Handler wraps a DNS Handler so that requests are traced.
type Handler struct {
	dns.Handler
}

// WrapHandler creates a new, wrapped DNS handler.
func WrapHandler(handler dns.Handler) *Handler {
	return &Handler{
		Handler: handler,
	}
}

// ServeDNS dispatches requests to the underlying Handler. All requests will be
// traced.
func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	span, _ := startSpan(context.Background(), r.Opcode)
	rw := &responseWriter{ResponseWriter: w}
	h.Handler.ServeDNS(rw, r)
	span.FinishWithOptionsExt(tracer.WithError(rw.err))
}

type responseWriter struct {
	dns.ResponseWriter
	err error
}

// WriteMsg writes the message to the response writer. If a non-success rcode is
// set the error in the struct will be non-nil.
func (rw *responseWriter) WriteMsg(msg *dns.Msg) error {
	if msg.Rcode != dns.RcodeSuccess {
		rw.err = errors.New(dns.RcodeToString[msg.Rcode])
	}
	return rw.ResponseWriter.WriteMsg(msg)
}

// Exchange calls dns.Exchange and traces the request.
func Exchange(m *dns.Msg, addr string) (r *dns.Msg, err error) {
	span, _ := startSpan(context.Background(), m.Opcode)
	r, err = dns.Exchange(m, addr)
	span.FinishWithOptionsExt(tracer.WithError(err))
	return r, err
}

// ExchangeConn calls dns.ExchangeConn and traces the request.
func ExchangeConn(c net.Conn, m *dns.Msg) (r *dns.Msg, err error) {
	span, _ := startSpan(context.Background(), m.Opcode)
	r, err = dns.ExchangeConn(c, m)
	span.FinishWithOptionsExt(tracer.WithError(err))
	return r, err
}

// ExchangeContext calls dns.ExchangeContext and traces the request.
func ExchangeContext(ctx context.Context, m *dns.Msg, addr string) (r *dns.Msg, err error) {
	span, ctx := startSpan(ctx, m.Opcode)
	r, err = dns.ExchangeContext(ctx, m, addr)
	span.FinishWithOptionsExt(tracer.WithError(err))
	return r, err
}

// A Client wraps a DNS Client so that requests are traced.
type Client struct {
	*dns.Client
}

// Exchange calls the underlying Client.Exchange and traces the request.
func (c *Client) Exchange(m *dns.Msg, addr string) (r *dns.Msg, rtt time.Duration, err error) {
	span, _ := startSpan(context.Background(), m.Opcode)
	r, rtt, err = c.Client.Exchange(m, addr)
	span.FinishWithOptionsExt(tracer.WithError(err))
	return r, rtt, err
}

// ExchangeContext calls the underlying Client.ExchangeContext and traces the request.
func (c *Client) ExchangeContext(ctx context.Context, m *dns.Msg, addr string) (r *dns.Msg, rtt time.Duration, err error) {
	span, ctx := startSpan(ctx, m.Opcode)
	r, rtt, err = c.Client.ExchangeContext(ctx, m, addr)
	span.FinishWithOptionsExt(tracer.WithError(err))
	return r, rtt, err
}

func startSpan(ctx context.Context, opcode int) (ddtrace.Span, context.Context) {
	return tracer.StartSpanFromContext(ctx, "dns.request",
		tracer.ServiceName("dns"),
		tracer.ResourceName(dns.OpcodeToString[opcode]),
		tracer.SpanType(ext.SpanTypeDNS))
}
