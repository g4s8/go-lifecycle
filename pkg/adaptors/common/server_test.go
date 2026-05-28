package common_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/g4s8/go-lifecycle/pkg/adaptors/common"
	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
)

// Compile-time interface check.
var _ lifecycle.Runner = (*common.HttpRunner)(nil)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())
	return port
}

func TestHttpRunner_Run_ContextCancel(t *testing.T) {
	port := freePort(t)
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port)}
	runner := common.NewHttpRunner(srv)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for runner to stop")
	}
}
