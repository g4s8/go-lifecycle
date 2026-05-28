package prometheus_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prometheusadaptor "github.com/g4s8/go-lifecycle/pkg/adaptors/prometheus"
	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
)

// Compile-time interface check.
var _ lifecycle.Runner = (*prometheusadaptor.Server)(nil)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())
	return port
}

func TestServer_Run_MetricsEndpoint(t *testing.T) {
	port := freePort(t)
	registry := prometheus.NewRegistry()

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_requests_total",
		Help: "Total test requests",
	}, []string{"method"})
	require.NoError(t, registry.Register(counter))
	counter.WithLabelValues("GET").Add(42)

	srv := prometheusadaptor.NewServer(registry, port, "/metrics")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.Run(ctx) //nolint:errcheck

	url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)
	var resp *http.Response
	var err error
	for i := 0; i < 50; i++ {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "test_requests_total")
}

func TestServer_Run_ContextCancel(t *testing.T) {
	port := freePort(t)
	registry := prometheus.NewRegistry()
	srv := prometheusadaptor.NewServer(registry, port, "/metrics")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for server to stop")
	}
}
