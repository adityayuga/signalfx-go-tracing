package grpc

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"

	"github.com/adityayuga/signalfx-go-tracing/ddtrace/ext"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/mocktracer"
)

func TestServerStatsHandler(t *testing.T) {
	assert := assert.New(t)

	serviceName := "grpc-service"
	statsHandler := NewServerStatsHandler(WithServiceName(serviceName))
	server, err := newServerStatsHandlerTestServer(statsHandler)
	if err != nil {
		t.Fatalf("failed to start test server: %s", err)
	}
	defer server.Close()

	mt := mocktracer.Start()
	defer mt.Stop()
	_, err = server.client.Ping(context.Background(), &FixtureRequest{Name: "name"})
	assert.NoError(err)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Zero(span.ParentID())
	assert.NotZero(span.StartTime())
	assert.True(span.FinishTime().After(span.StartTime()))
	assert.Equal("grpc.server", span.OperationName())
	assert.Equal(map[string]interface{}{
		"span.type":     ext.AppTypeRPC,
		"grpc.code":     codes.OK.String(),
		"service.name":  serviceName,
		"resource.name": "/grpc.Fixture/Ping",
		"grpc.method":   "/grpc.Fixture/Ping",
	}, span.Tags())
}

func newServerStatsHandlerTestServer(statsHandler stats.Handler) (*rig, error) {
	server := grpc.NewServer(grpc.StatsHandler(statsHandler))
	fixtureServer := new(fixtureServer)
	RegisterFixtureServer(server, fixtureServer)

	li, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	_, port, _ := net.SplitHostPort(li.Addr().String())
	go server.Serve(li)

	conn, err := grpc.Dial(li.Addr().String(), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("error dialing: %s", err)
	}
	return &rig{
		fixtureServer: fixtureServer,
		listener:      li,
		port:          port,
		server:        server,
		conn:          conn,
		client:        NewFixtureClient(conn),
	}, nil
}
