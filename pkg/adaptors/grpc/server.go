package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
)

const gracefulShutdownTimeout = 5 * time.Second

// GRPCRunner wraps a grpc.Server and implements the lifecycle.Runner interface.
type GRPCRunner struct {
	srv  *grpc.Server
	addr string
}

// NewGRPCRunner creates a new GRPCRunner for the given address and server.
func NewGRPCRunner(addr string, srv *grpc.Server) *GRPCRunner {
	return &GRPCRunner{srv: srv, addr: addr}
}

func (g *GRPCRunner) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", g.addr)
	if err != nil {
		return fmt.Errorf("listen address: %w", err)
	}

	errCh := make(chan error, 1)
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		if err := g.srv.Serve(l); err != nil && err != grpc.ErrServerStopped {
			errCh <- fmt.Errorf("serve grpc: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		stopped := make(chan struct{})
		go func() {
			g.srv.GracefulStop()
			close(stopped)
		}()
		t := time.NewTimer(gracefulShutdownTimeout)
		defer t.Stop()
		select {
		case <-stopped:
		case <-t.C:
			g.srv.Stop()
		}
	case err := <-errCh:
		<-doneCh
		return err
	}

	<-doneCh

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
