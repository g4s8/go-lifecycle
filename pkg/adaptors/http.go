package adaptors

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/types"
	"github.com/pkg/errors"
)

// HTTPService is an adapter for http.Server to implement lifecycle hooks.
type HTTPService struct {
	srv *http.Server
}

// NewHTTPService creates new HTTPService adaptor.
func NewHTTPService(srv *http.Server) *HTTPService {
	return &HTTPService{srv: srv}
}

// RegisterLifecycle registers this service in lifecycle manager.
func (s *HTTPService) RegisterLifecycle(name string, lf LifecycleRegistry) {
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

func (s *HTTPService) Start(ctx context.Context, errCh chan<- error) error {
	addr := s.srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	go func() {
		if err := s.srv.Serve(ln); err != nil {
			if err != http.ErrServerClosed {
				errCh <- errors.Wrap(err, "serve")
				return
			}
		}
	}()
	return nil
}

func (s *HTTPService) Stop(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "shutdown")
	}
	return nil
}
