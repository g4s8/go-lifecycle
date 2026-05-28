package grpc_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	grpcadaptor "github.com/g4s8/go-lifecycle/pkg/adaptors/grpc"
	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
)

// Compile-time interface check.
var _ lifecycle.Runner = (*grpcadaptor.GRPCRunner)(nil)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())
	return port
}

func TestGRPCRunner_Run_ContextCancel(t *testing.T) {
	port := freePort(t)
	srv := grpc.NewServer()

	runner := grpcadaptor.NewGRPCRunner(fmt.Sprintf("127.0.0.1:%d", port), srv)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for gRPC runner to stop")
	}
}

func TestGRPCRunner_Run_ListenError(t *testing.T) {
	port := freePort(t)

	// Occupy the port
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	require.NoError(t, err)
	defer ln.Close()

	srv := grpc.NewServer()
	runner := grpcadaptor.NewGRPCRunner(fmt.Sprintf("127.0.0.1:%d", port), srv)

	err = runner.Run(context.Background())
	require.Error(t, err)
}

func TestGRPCRunner_HealthCheck(t *testing.T) {
	port := freePort(t)
	srv := grpc.NewServer()

	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	runner := grpcadaptor.NewGRPCRunner(fmt.Sprintf("127.0.0.1:%d", port), srv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	go func() {
		close(ready)
		runner.Run(ctx) //nolint:errcheck
	}()
	<-ready

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	var conn *grpc.ClientConn
	var connErr error
	for i := 0; i < 50; i++ {
		//nolint:staticcheck // grpc.Dial is deprecated in newer versions, but required for v1.53
		conn, connErr = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, connErr)
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	var resp *grpc_health_v1.HealthCheckResponse
	var checkErr error
	for i := 0; i < 50; i++ {
		resp, checkErr = client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
		if checkErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, checkErr)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
}
