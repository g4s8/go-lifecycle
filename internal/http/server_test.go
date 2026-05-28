package http_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"context"

	internal_http "github.com/g4s8/go-lifecycle/internal/http"
)

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

func TestNewServer_Defaults(t *testing.T) {
	srv := internal_http.NewServer("127.0.0.1", 8080)
	assert.Equal(t, "127.0.0.1:8080", srv.Addr)
	assert.Equal(t, internal_http.DefaultReadTimeout, srv.ReadTimeout)
}

func TestNewServer_WithMux(t *testing.T) {
	mux := http.NewServeMux()
	srv := internal_http.NewServer("", 8080, internal_http.WithMux(mux))
	assert.Equal(t, mux, srv.Handler)
}

func TestNewServer_WithReadTimeout(t *testing.T) {
	timeout := 10 * time.Second
	srv := internal_http.NewServer("", 8080, internal_http.WithReadTimeout(timeout))
	assert.Equal(t, timeout, srv.ReadTimeout)
}

func TestRunServer_ContextCancelShutdown(t *testing.T) {
	port := freePort(t)
	srv := internal_http.NewServer("127.0.0.1", port)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- internal_http.RunServer(ctx, srv) }()

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for server shutdown")
	}
}

func TestRunServer_HandlerResponds(t *testing.T) {
	port := freePort(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := internal_http.NewServer("127.0.0.1", port, internal_http.WithMux(mux))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	go func() {
		close(ready)
		internal_http.RunServer(ctx, srv) //nolint:errcheck
	}()
	<-ready

	// Give server a moment to start
	var resp *http.Response
	var err error
	for i := 0; i < 20; i++ {
		resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRunServer_ListenError(t *testing.T) {
	port := freePort(t)

	// Occupy the port
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	require.NoError(t, err)
	defer ln.Close()

	srv := internal_http.NewServer("127.0.0.1", port)
	err = internal_http.RunServer(context.Background(), srv)
	require.Error(t, err)
}
