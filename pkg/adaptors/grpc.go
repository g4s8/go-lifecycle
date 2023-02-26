package adaptors

import (
	"context"
	"net"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// GRPCService is a gRPC server adapter for lifecycle manager.
type GRPCService struct {
	addr string
	srv  *grpc.Server

	l net.Listener
}

// NewGRPCService creates new gRPC server adapter.
func NewGRPCService(addr string, srv *grpc.Server) *GRPCService {
	return &GRPCService{addr: addr, srv: srv}
}

// RegisterLifecycle registers service in lifecycle manager.
func (s *GRPCService) RegisterLifecycle(name string, lf LifecycleRegistry) {
	lf.RegisterService(types.ServiceConfig{
		Name:         name,
		StartupHook:  s.Start,
		ShutdownHook: s.Stop,
		RestartPolicy: types.ServiceRestartPolicy{
			RestartOnFailure: true,
			RestartCount:     3,
			RestartDelay:     time.Millisecond * 200,
		},
	})
}

func (s *GRPCService) Start(ctx context.Context, errCh chan<- error) error {
	var err error
	s.l, err = net.Listen("tcp", s.addr)
	if err != nil {
		return errors.Wrap(err, "listed address")
	}
	go func() {
		if err := s.srv.Serve(s.l); err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return
			}
			errCh <- errors.Wrap(err, "serve grpc")
		}
	}()
	return nil
}

func (s *GRPCService) Stop(ctx context.Context) error {
	stopCh := make(chan struct{}, 1)
	go func() {
		s.srv.GracefulStop()
		close(stopCh)
	}()
	select {
	case <-ctx.Done():
		s.srv.Stop()
		return ctx.Err()
	case <-stopCh:
	}
	if err := s.l.Close(); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			return errors.Wrap(err, "close listener")
		}
	}
	return nil
}
